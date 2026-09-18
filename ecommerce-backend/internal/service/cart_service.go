package service

import (
	"context"
	"errors"
	"fmt"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// ErrInsufficientStock means the requested quantity exceeds the product's
// available stock.
var ErrInsufficientStock = errors.New("insufficient stock")

// CartService implements the shopping cart's business rules: stock checks
// on add/update, and row-level ownership — a user can only ever see or
// modify their own cart's items.
type CartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

// NewCartService builds a CartService with its dependencies.
func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) *CartService {
	return &CartService{cartRepo: cartRepo, productRepo: productRepo}
}

// Get returns userID's cart, creating an empty one if they don't have one
// yet.
func (s *CartService) Get(ctx context.Context, userID uint) (*model.CartResponse, error) {
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: get cart: %w", err))
	}
	items, err := s.cartRepo.ListItems(ctx, cart.ID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: get cart: %w", err))
	}
	return buildCartResponse(items), nil
}

// AddItem adds req.Quantity of a product to userID's cart, incrementing the
// existing line if that product is already in the cart. The combined
// quantity must not exceed the product's stock.
func (s *CartService) AddItem(ctx context.Context, userID uint, req model.AddCartItemRequest) (*model.CartResponse, error) {
	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: add cart item: %w", err))
	}
	if product == nil {
		return nil, apperror.Validation("Produk tidak ditemukan", []string{"product_id: does not exist"})
	}

	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: add cart item: %w", err))
	}

	existing, err := s.cartRepo.FindItemByProductID(ctx, cart.ID, req.ProductID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: add cart item: %w", err))
	}

	targetQuantity := req.Quantity
	if existing != nil {
		targetQuantity += existing.Quantity
	}
	if targetQuantity > product.Stock {
		return nil, apperror.Conflict("Stok tidak mencukupi", ErrInsufficientStock)
	}

	if existing != nil {
		existing.Quantity = targetQuantity
		if err := s.cartRepo.UpdateItemQuantity(ctx, existing); err != nil {
			return nil, apperror.Internal(fmt.Errorf("service: add cart item: %w", err))
		}
	} else {
		item := &model.CartItem{CartID: cart.ID, ProductID: req.ProductID, Quantity: req.Quantity}
		if err := s.cartRepo.CreateItem(ctx, item); err != nil {
			return nil, apperror.Internal(fmt.Errorf("service: add cart item: %w", err))
		}
	}

	return s.Get(ctx, userID)
}

// UpdateItemQuantity sets an item's quantity, after verifying it belongs to
// userID's own cart and that the new quantity fits the product's stock.
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID, itemID uint, req model.UpdateCartItemRequest) (*model.CartResponse, error) {
	item, err := s.findOwnedItem(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, apperror.NotFound("Item cart tidak ditemukan", nil)
	}
	if req.Quantity > item.Product.Stock {
		return nil, apperror.Conflict("Stok tidak mencukupi", ErrInsufficientStock)
	}

	item.Quantity = req.Quantity
	if err := s.cartRepo.UpdateItemQuantity(ctx, item); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update cart item: %w", err))
	}
	return s.Get(ctx, userID)
}

// RemoveItem deletes an item after verifying it belongs to userID's own
// cart.
func (s *CartService) RemoveItem(ctx context.Context, userID, itemID uint) (*model.CartResponse, error) {
	item, err := s.findOwnedItem(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, apperror.NotFound("Item cart tidak ditemukan", nil)
	}

	if err := s.cartRepo.DeleteItem(ctx, item.ID); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: remove cart item: %w", err))
	}
	return s.Get(ctx, userID)
}

// findOwnedItem fetches itemID and verifies it belongs to userID's own
// cart — the row-level ownership check called out in the RBAC section of
// .claude/CLAUDE.md, separate from the route-level RequirePermission
// gate. A mismatch reports back as "not found" (nil, nil) rather than
// "forbidden", so a user probing other people's item IDs can't tell which
// ones actually exist.
func (s *CartService) findOwnedItem(ctx context.Context, userID, itemID uint) (*model.CartItem, error) {
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: find owned cart item: %w", err))
	}

	item, err := s.cartRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: find owned cart item: %w", err))
	}
	if item == nil || item.CartID != cart.ID {
		return nil, nil
	}
	return item, nil
}

func buildCartResponse(items []model.CartItem) *model.CartResponse {
	response := &model.CartResponse{Items: make([]model.CartItemResponse, 0, len(items))}
	for _, item := range items {
		subtotal := item.Product.Price * float64(item.Quantity)
		response.Items = append(response.Items, model.CartItemResponse{
			ID:       item.ID,
			Product:  *item.Product,
			Quantity: item.Quantity,
			Subtotal: subtotal,
		})
		response.Total += subtotal
	}
	return response
}
