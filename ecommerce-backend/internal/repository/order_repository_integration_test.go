//go:build integration

package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/testdb"
)

// TestOrderRepository_Checkout_ConcurrentRequestsDoNotOversell reproduces,
// against a real Postgres container, the exact race condition this project
// was previously verified against only manually (two concurrent `curl`
// requests, see .claude/CLAUDE.md Current Focus). Two users each have 1
// unit of a stock=1 product in their cart; checking out at the same time
// must leave exactly one order created and stock at 0, never -1 — proving
// the SELECT ... FOR UPDATE row lock in Checkout actually serializes them.
func TestOrderRepository_Checkout_ConcurrentRequestsDoNotOversell(t *testing.T) {
	db := testdb.New(t)
	orderRepo := repository.NewOrderRepository(db)

	productID := seedCategoryAndProduct(t, db, 1)
	userA := seedUserWithCartItem(t, db, "user-a@example.com", productID, 1)
	userB := seedUserWithCartItem(t, db, "user-b@example.com", productID, 1)

	var wg sync.WaitGroup
	results := make([]error, 2)
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, err := orderRepo.Checkout(context.Background(), userA, "Jl. A No. 1", "")
		results[0] = err
	}()
	go func() {
		defer wg.Done()
		_, err := orderRepo.Checkout(context.Background(), userB, "Jl. B No. 1", "")
		results[1] = err
	}()
	wg.Wait()

	successes, failures := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, repository.ErrInsufficientStock):
			failures++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successes != 1 || failures != 1 {
		t.Fatalf("expected exactly 1 success and 1 failure, got %d successes and %d failures", successes, failures)
	}

	var product model.Product
	if err := db.First(&product, productID).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if product.Stock != 0 {
		t.Errorf("Stock = %d, want 0 (oversold or under-decremented)", product.Stock)
	}

	var orderCount int64
	if err := db.Model(&model.Order{}).Count(&orderCount).Error; err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if orderCount != 1 {
		t.Errorf("orders created = %d, want exactly 1", orderCount)
	}
}

// TestOrderRepository_Checkout_SnapshotsProductName is a regression test:
// OrderItem.ProductName must be captured at checkout time and stay
// unchanged even if the product is renamed afterwards — previously only
// PriceAtPurchase was snapshotted, so a renamed/deleted product silently
// rewrote historical order data (see .claude/CLAUDE.md Known Issues).
func TestOrderRepository_Checkout_SnapshotsProductName(t *testing.T) {
	db := testdb.New(t)
	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)

	productID := seedCategoryAndProduct(t, db, 5)
	userID := seedUserWithCartItem(t, db, "renamer@example.com", productID, 1)

	order, err := orderRepo.Checkout(context.Background(), userID, "Jl. Snapshot No. 1", "")
	if err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if order.Items[0].ProductName != "Kopi Susu" {
		t.Fatalf("ProductName = %q, want %q", order.Items[0].ProductName, "Kopi Susu")
	}

	product, err := productRepo.FindByID(context.Background(), productID)
	if err != nil || product == nil {
		t.Fatalf("FindByID: %v", err)
	}
	product.Name = "Renamed After Purchase"
	if err := productRepo.Update(context.Background(), product); err != nil {
		t.Fatalf("Update: %v", err)
	}

	var reloaded model.OrderItem
	if err := db.First(&reloaded, order.Items[0].ID).Error; err != nil {
		t.Fatalf("reload order item: %v", err)
	}
	if reloaded.ProductName != "Kopi Susu" {
		t.Errorf("ProductName after product rename = %q, want unchanged %q", reloaded.ProductName, "Kopi Susu")
	}
}

// TestOrderRepository_Checkout_EmptyCartReturnsError checks the guard that
// exists purely to be race-safe with the concurrent case above: a user with
// no cart items at all must never reach the stock-locking logic.
func TestOrderRepository_Checkout_EmptyCartReturnsError(t *testing.T) {
	db := testdb.New(t)
	orderRepo := repository.NewOrderRepository(db)

	user := model.User{Name: "No Cart", Email: "no-cart@example.com", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.Cart{UserID: user.ID}).Error; err != nil {
		t.Fatalf("seed empty cart: %v", err)
	}

	_, err := orderRepo.Checkout(context.Background(), user.ID, "Jl. Kosong No. 1", "")
	if !errors.Is(err, repository.ErrEmptyCart) {
		t.Fatalf("Checkout error = %v, want ErrEmptyCart", err)
	}
}
