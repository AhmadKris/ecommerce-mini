//go:build integration

package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/testdb"
)

// TestOrderRepository_Checkout_ConcurrentCheckoutsDoNotOverredeemPromo is
// the promo-code analog of the stock/inventory race tests: two users each
// have a valid cart and both redeem the same promo code, which has
// usage_limit=1. Checking out at the same time must leave exactly one
// order with the discount applied and the promotion's used_count at
// exactly 1 (never 2) — proving lockAndRedeemPromotion's row lock
// serializes concurrent redemptions the same way lockAndReserveStock does
// for product stock. The loser's checkout must fail entirely (whole
// transaction rolled back), not silently succeed without the discount.
func TestOrderRepository_Checkout_ConcurrentCheckoutsDoNotOverredeemPromo(t *testing.T) {
	db := testdb.New(t)
	orderRepo := repository.NewOrderRepository(db)

	productID := seedCategoryAndProduct(t, db, 10)
	userA := seedUserWithCartItem(t, db, "promo-a@example.com", productID, 1)
	userB := seedUserWithCartItem(t, db, "promo-b@example.com", productID, 1)

	promotion := model.Promotion{
		Code: "RACE10", Type: model.PromotionTypePercentage, Value: 10,
		UsageLimit: 1, StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(time.Hour),
		Status: model.PromotionStatusActive,
	}
	if err := db.Create(&promotion).Error; err != nil {
		t.Fatalf("seed promotion: %v", err)
	}

	var wg sync.WaitGroup
	results := make([]error, 2)
	wg.Add(2)

	checkout := func(i int, userID uint) {
		defer wg.Done()
		_, err := orderRepo.Checkout(context.Background(), userID, "Jl. Promo Race No. 1", "RACE10")
		results[i] = err
	}
	go checkout(0, userA)
	go checkout(1, userB)
	wg.Wait()

	successes, failures := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, repository.ErrPromotionExhausted):
			failures++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("expected exactly 1 success and 1 failure, got %d successes and %d failures: %v", successes, failures, results)
	}

	var reloadedPromotion model.Promotion
	if err := db.First(&reloadedPromotion, promotion.ID).Error; err != nil {
		t.Fatalf("reload promotion: %v", err)
	}
	if reloadedPromotion.UsedCount != 1 {
		t.Errorf("UsedCount = %d, want exactly 1 (over-redeemed or not counted)", reloadedPromotion.UsedCount)
	}

	var orderCount int64
	if err := db.Model(&model.Order{}).Where("promotion_id = ?", promotion.ID).Count(&orderCount).Error; err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if orderCount != 1 {
		t.Errorf("orders with this promotion = %d, want exactly 1", orderCount)
	}
}
