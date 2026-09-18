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

// TestInventoryRepository_Adjust_ConcurrentOutAdjustmentsDoNotGoNegative is
// the inventory-adjustment analog of the checkout race-condition test:
// two concurrent "out" adjustments for a product with stock=1, each
// requesting 1 unit, must leave exactly one success and stock at 0, never
// -1 — proving the same SELECT ... FOR UPDATE lock pattern used in
// OrderRepository.Checkout also holds here.
func TestInventoryRepository_Adjust_ConcurrentOutAdjustmentsDoNotGoNegative(t *testing.T) {
	db := testdb.New(t)
	inventoryRepo := repository.NewInventoryRepository(db)

	productID := seedCategoryAndProduct(t, db, 1)
	actorID := seedUser(t, db, "adjuster@example.com")

	var wg sync.WaitGroup
	results := make([]error, 2)
	wg.Add(2)

	adjust := func(i int) {
		defer wg.Done()
		_, err := inventoryRepo.Adjust(context.Background(), actorID, model.AdjustInventoryRequest{
			ProductID: productID, Type: model.InventoryMovementTypeOut, Quantity: 1, Reason: "concurrent test",
		})
		results[i] = err
	}
	go adjust(0)
	go adjust(1)
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
}
