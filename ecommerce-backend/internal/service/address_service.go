package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// AddressService implements the shipping address book: row-level ownership
// (a user can only ever see/modify their own addresses) and the
// at-most-one-default invariant.
type AddressService struct {
	addressRepo repository.AddressRepository
}

// NewAddressService builds an AddressService with its dependencies.
func NewAddressService(addressRepo repository.AddressRepository) *AddressService {
	return &AddressService{addressRepo: addressRepo}
}

// List returns userID's saved addresses, default first.
func (s *AddressService) List(ctx context.Context, userID uint) ([]model.Address, error) {
	addresses, err := s.addressRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: list addresses: %w", err))
	}
	return addresses, nil
}

// Create adds a new address to userID's address book.
func (s *AddressService) Create(ctx context.Context, userID uint, req model.CreateAddressRequest) (*model.Address, error) {
	address := &model.Address{
		UserID:        userID,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		AddressLine:   req.AddressLine,
		City:          req.City,
		Province:      req.Province,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
	}
	if err := s.addressRepo.Create(ctx, address); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: create address: %w", err))
	}
	return address, nil
}

// Update replaces an existing address's fields, after verifying it belongs
// to userID.
func (s *AddressService) Update(ctx context.Context, userID uint, id uint, req model.UpdateAddressRequest) (*model.Address, error) {
	address, err := s.findOwnedAddress(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if address == nil {
		return nil, apperror.NotFound("Alamat tidak ditemukan", nil)
	}

	address.RecipientName = req.RecipientName
	address.Phone = req.Phone
	address.AddressLine = req.AddressLine
	address.City = req.City
	address.Province = req.Province
	address.PostalCode = req.PostalCode
	address.IsDefault = req.IsDefault

	if err := s.addressRepo.Update(ctx, address); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update address: %w", err))
	}
	return address, nil
}

// Delete removes an address after verifying it belongs to userID.
func (s *AddressService) Delete(ctx context.Context, userID uint, id uint) error {
	address, err := s.findOwnedAddress(ctx, userID, id)
	if err != nil {
		return err
	}
	if address == nil {
		return apperror.NotFound("Alamat tidak ditemukan", nil)
	}

	if err := s.addressRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Alamat tidak ditemukan", nil)
		}
		return apperror.Internal(fmt.Errorf("service: delete address: %w", err))
	}
	return nil
}

// findOwnedAddress fetches id and verifies it belongs to userID — same
// row-level ownership pattern as CartService.findOwnedItem: a mismatch
// reports back as (nil, nil), which callers turn into NotFound (not
// Forbidden), so a user probing other people's address IDs can't tell which
// ones actually exist.
func (s *AddressService) findOwnedAddress(ctx context.Context, userID, id uint) (*model.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: find owned address: %w", err))
	}
	if address == nil || address.UserID != userID {
		return nil, nil
	}
	return address, nil
}
