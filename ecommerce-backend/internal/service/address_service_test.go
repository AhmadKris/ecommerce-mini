package service

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
)

type fakeAddressRepo struct {
	addresses map[uint]*model.Address
	nextID    uint
}

func newFakeAddressRepo() *fakeAddressRepo {
	return &fakeAddressRepo{addresses: map[uint]*model.Address{}, nextID: 1}
}

func (r *fakeAddressRepo) Create(_ context.Context, address *model.Address) error {
	if address.IsDefault {
		for _, other := range r.addresses {
			other.IsDefault = false
		}
	}
	address.ID = r.nextID
	r.nextID++
	r.addresses[address.ID] = address
	return nil
}

func (r *fakeAddressRepo) Update(_ context.Context, address *model.Address) error {
	if address.IsDefault {
		for id, other := range r.addresses {
			if id != address.ID {
				other.IsDefault = false
			}
		}
	}
	r.addresses[address.ID] = address
	return nil
}

func (r *fakeAddressRepo) Delete(_ context.Context, id uint) error {
	if _, ok := r.addresses[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.addresses, id)
	return nil
}

func (r *fakeAddressRepo) FindByID(_ context.Context, id uint) (*model.Address, error) {
	return r.addresses[id], nil
}

func (r *fakeAddressRepo) ListByUserID(_ context.Context, userID uint) ([]model.Address, error) {
	var result []model.Address
	for _, address := range r.addresses {
		if address.UserID == userID {
			result = append(result, *address)
		}
	}
	return result, nil
}

func TestAddressService_Update_RejectsOtherUsersAddress(t *testing.T) {
	repo := newFakeAddressRepo()
	svc := NewAddressService(repo)

	owned, err := svc.Create(context.Background(), 1, model.CreateAddressRequest{
		RecipientName: "Owner", Phone: "0800", AddressLine: "Jl. Owner No. 1", City: "Jakarta", Province: "DKI", PostalCode: "10110",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = svc.Update(context.Background(), 2, owned.ID, model.UpdateAddressRequest{
		RecipientName: "Intruder", Phone: "0800", AddressLine: "Jl. Intruder No. 1", City: "Bandung", Province: "Jabar", PostalCode: "40111",
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("Update by non-owner: error = %v, want apperror NotFound", err)
	}

	err = svc.Delete(context.Background(), 2, owned.ID)
	appErr, ok = apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("Delete by non-owner: error = %v, want apperror NotFound", err)
	}

	reloaded, _ := repo.FindByID(context.Background(), owned.ID)
	if reloaded == nil || reloaded.RecipientName != "Owner" {
		t.Fatalf("owner's address changed after rejected attempts: %+v", reloaded)
	}
}

func TestAddressService_Create_OnlyOneDefaultAtATime(t *testing.T) {
	repo := newFakeAddressRepo()
	svc := NewAddressService(repo)

	first, err := svc.Create(context.Background(), 1, model.CreateAddressRequest{
		RecipientName: "Home", Phone: "0800", AddressLine: "Jl. Rumah No. 1", City: "Jakarta", Province: "DKI", PostalCode: "10110",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("Create first returned error: %v", err)
	}

	_, err = svc.Create(context.Background(), 1, model.CreateAddressRequest{
		RecipientName: "Office", Phone: "0800", AddressLine: "Jl. Kantor No. 1", City: "Jakarta", Province: "DKI", PostalCode: "10120",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("Create second returned error: %v", err)
	}

	addresses, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	defaultCount := 0
	for _, addr := range addresses {
		if addr.IsDefault {
			defaultCount++
		}
	}
	if defaultCount != 1 {
		t.Fatalf("expected exactly 1 default address, got %d: %+v", defaultCount, addresses)
	}

	reloadedFirst, _ := repo.FindByID(context.Background(), first.ID)
	if reloadedFirst.IsDefault {
		t.Error("first address should no longer be default after second was created as default")
	}
}
