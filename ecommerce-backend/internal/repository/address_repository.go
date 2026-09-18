package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// AddressRepository persists a user's saved shipping addresses.
type AddressRepository interface {
	Create(ctx context.Context, address *model.Address) error
	Update(ctx context.Context, address *model.Address) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Address, error)
	ListByUserID(ctx context.Context, userID uint) ([]model.Address, error)
}

type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository builds an AddressRepository backed by db.
func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, address *model.Address) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if address.IsDefault {
			if err := tx.Model(&model.Address{}).
				Where("user_id = ?", address.UserID).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(address).Error
	})
}

func (r *addressRepository) Update(ctx context.Context, address *model.Address) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if address.IsDefault {
			if err := tx.Model(&model.Address{}).
				Where("user_id = ? AND id != ?", address.UserID, address.ID).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(address).Error
	})
}

func (r *addressRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Address{}, id)
	if result.Error != nil {
		return fmt.Errorf("repository: delete address: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository: delete address: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *addressRepository) FindByID(ctx context.Context, id uint) (*model.Address, error) {
	var address model.Address
	err := r.db.WithContext(ctx).First(&address, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find address by id: %w", err)
	}
	return &address, nil
}

func (r *addressRepository) ListByUserID(ctx context.Context, userID uint) ([]model.Address, error) {
	var addresses []model.Address
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error
	if err != nil {
		return nil, fmt.Errorf("repository: list addresses: %w", err)
	}
	return addresses, nil
}
