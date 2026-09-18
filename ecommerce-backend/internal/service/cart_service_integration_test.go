//go:build integration

package service_test

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/internal/testdb"
)

func seedProductForCart(t *testing.T, db *gorm.DB, stock int) uint {
	t.Helper()
	category := model.Category{Name: "Minuman", Slug: "minuman-cart-test"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	product := model.Product{Name: "Kopi Susu", Slug: "kopi-susu-cart-test", Price: 18000, Stock: stock, CategoryID: category.ID}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
	return product.ID
}

func seedUser(t *testing.T, db *gorm.DB, email string) uint {
	t.Helper()
	user := model.User{Name: "Test User", Email: email, PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user.ID
}

// TestCartService_UpdateItemQuantity_RejectsOtherUsersItem is the automated
// version of the ownership check previously only verified manually (two
// real accounts via curl, see .claude/CLAUDE.md Known Issues): user B must
// not be able to update or delete an item in user A's cart, and the
// response must be NotFound (not Forbidden) so B can't distinguish "not
// yours" from "doesn't exist" — see CartService.findOwnedItem's doc comment.
func TestCartService_UpdateItemQuantity_RejectsOtherUsersItem(t *testing.T) {
	db := testdb.New(t)
	cartRepo := repository.NewCartRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartService := service.NewCartService(cartRepo, productRepo)

	productID := seedProductForCart(t, db, 10)
	userA := seedUser(t, db, "owner@example.com")
	userB := seedUser(t, db, "intruder@example.com")

	ctx := context.Background()
	cartA, err := cartService.AddItem(ctx, userA, model.AddCartItemRequest{ProductID: productID, Quantity: 2})
	if err != nil {
		t.Fatalf("user A add item: %v", err)
	}
	itemID := cartA.Items[0].ID

	_, err = cartService.UpdateItemQuantity(ctx, userB, itemID, model.UpdateCartItemRequest{Quantity: 5})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("UpdateItemQuantity by non-owner: error = %v, want apperror NotFound", err)
	}

	_, err = cartService.RemoveItem(ctx, userB, itemID)
	appErr, ok = apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("RemoveItem by non-owner: error = %v, want apperror NotFound", err)
	}

	// The item must still be exactly as user A left it — the rejected
	// attempts by user B must not have mutated it.
	cartAAfter, err := cartService.Get(ctx, userA)
	if err != nil {
		t.Fatalf("reload user A cart: %v", err)
	}
	if len(cartAAfter.Items) != 1 || cartAAfter.Items[0].Quantity != 2 {
		t.Fatalf("user A's cart changed after B's rejected attempts: %+v", cartAAfter.Items)
	}
}
