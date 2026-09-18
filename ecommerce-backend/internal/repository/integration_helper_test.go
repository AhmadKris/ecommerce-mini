//go:build integration

package repository_test

import (
	"testing"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// seedCategoryAndProduct inserts one category and one product with the
// given stock, returning the product's ID.
func seedCategoryAndProduct(t *testing.T, db *gorm.DB, stock int) uint {
	t.Helper()

	category := model.Category{Name: "Minuman", Slug: "minuman-test"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}

	product := model.Product{
		Name:       "Kopi Susu",
		Slug:       "kopi-susu-test",
		Price:      18000,
		Stock:      stock,
		CategoryID: category.ID,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
	return product.ID
}

// seedUserWithCartItem inserts a user with a cart containing one line item
// (productID, quantity), returning the user's ID.
func seedUserWithCartItem(t *testing.T, db *gorm.DB, email string, productID uint, quantity int) uint {
	t.Helper()

	user := model.User{Name: "Test User", Email: email, PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	cart := model.Cart{UserID: user.ID}
	if err := db.Create(&cart).Error; err != nil {
		t.Fatalf("seed cart: %v", err)
	}

	item := model.CartItem{CartID: cart.ID, ProductID: productID, Quantity: quantity}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed cart item: %v", err)
	}

	return user.ID
}

// seedUser inserts a plain user (no cart), returning its ID.
func seedUser(t *testing.T, db *gorm.DB, email string) uint {
	t.Helper()

	user := model.User{Name: "Test User", Email: email, PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user.ID
}
