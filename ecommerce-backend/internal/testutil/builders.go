// Package testutil holds fixture builders and assertion helpers shared
// across this project's test files — the "what's the default valid X"
// definition for each entity lives here exactly once, instead of every
// test file hand-writing its own slightly-different struct literal.
// Importable (not _test.go) so it works from any package's tests.
package testutil

import "ecommerce-backend/internal/model"

// ProductOption customizes a fixture built by NewProduct.
type ProductOption func(*model.Product)

func WithProductID(id uint) ProductOption {
	return func(p *model.Product) { p.ID = id }
}

func WithProductName(name string) ProductOption {
	return func(p *model.Product) { p.Name = name }
}

func WithStock(stock int) ProductOption {
	return func(p *model.Product) { p.Stock = stock }
}

func WithPrice(price float64) ProductOption {
	return func(p *model.Product) { p.Price = price }
}

func WithProductCategoryID(categoryID uint) ProductOption {
	return func(p *model.Product) { p.CategoryID = categoryID }
}

// NewProduct returns a valid product fixture (in stock, priced, named),
// customized via options — e.g. NewProduct(WithStock(0)) for an
// out-of-stock case.
func NewProduct(opts ...ProductOption) *model.Product {
	product := &model.Product{
		ID:         1,
		Name:       "Kopi Susu Gula Aren",
		Slug:       "kopi-susu-gula-aren",
		Price:      18000,
		Stock:      10,
		CategoryID: 1,
	}
	for _, opt := range opts {
		opt(product)
	}
	return product
}

// CategoryOption customizes a fixture built by NewCategory.
type CategoryOption func(*model.Category)

func WithCategoryID(id uint) CategoryOption {
	return func(c *model.Category) { c.ID = id }
}

func WithCategoryName(name string) CategoryOption {
	return func(c *model.Category) { c.Name = name }
}

// NewCategory returns a valid category fixture, customized via options.
func NewCategory(opts ...CategoryOption) *model.Category {
	category := &model.Category{ID: 1, Name: "Minuman", Slug: "minuman"}
	for _, opt := range opts {
		opt(category)
	}
	return category
}

// UserOption customizes a fixture built by NewUser.
type UserOption func(*model.User)

func WithUserID(id uint) UserOption {
	return func(u *model.User) { u.ID = id }
}

func WithUserEmail(email string) UserOption {
	return func(u *model.User) { u.Email = email }
}

// NewUser returns a valid user fixture, customized via options.
func NewUser(opts ...UserOption) *model.User {
	user := &model.User{ID: 1, Name: "Test User", Email: "test@example.com", PasswordHash: "hash"}
	for _, opt := range opts {
		opt(user)
	}
	return user
}

// CartItemOption customizes a fixture built by NewCartItem.
type CartItemOption func(*model.CartItem)

func WithCartItemID(id uint) CartItemOption {
	return func(i *model.CartItem) { i.ID = id }
}

func WithCartItemQuantity(quantity int) CartItemOption {
	return func(i *model.CartItem) { i.Quantity = quantity }
}

func WithCartItemProduct(product *model.Product) CartItemOption {
	return func(i *model.CartItem) {
		i.ProductID = product.ID
		i.Product = product
	}
}

// NewCartItem returns a valid cart item fixture — quantity 1 of
// NewProduct()'s default product, in cart 1 — customized via options.
func NewCartItem(opts ...CartItemOption) *model.CartItem {
	product := NewProduct()
	item := &model.CartItem{ID: 1, CartID: 1, ProductID: product.ID, Quantity: 1, Product: product}
	for _, opt := range opts {
		opt(item)
	}
	return item
}
