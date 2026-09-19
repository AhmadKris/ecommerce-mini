package service

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

type fakeProductRepo struct {
	products map[uint]*model.Product
	nextID   uint
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{products: map[uint]*model.Product{}, nextID: 1}
}

func (r *fakeProductRepo) Create(_ context.Context, product *model.Product) error {
	product.ID = r.nextID
	r.nextID++
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) Update(_ context.Context, product *model.Product) error {
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) Delete(_ context.Context, id uint) error {
	if _, ok := r.products[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.products, id)
	return nil
}

func (r *fakeProductRepo) FindByID(_ context.Context, id uint) (*model.Product, error) {
	return r.products[id], nil
}

func (r *fakeProductRepo) FindBySlug(_ context.Context, slug string) (*model.Product, error) {
	for _, product := range r.products {
		if product.Slug == slug {
			return product, nil
		}
	}
	return nil, nil
}

func (r *fakeProductRepo) List(context.Context, repository.ProductFilter) ([]model.Product, int64, error) {
	return nil, 0, nil
}

type fakeCategoryRepo struct{}

func (fakeCategoryRepo) Create(context.Context, *model.Category) error { return nil }
func (fakeCategoryRepo) Update(context.Context, *model.Category) error { return nil }
func (fakeCategoryRepo) Delete(context.Context, uint) error            { return nil }

func (fakeCategoryRepo) FindByID(_ context.Context, id uint) (*model.Category, error) {
	return &model.Category{ID: id, Name: "Test Category", Slug: "test-category"}, nil
}

func (fakeCategoryRepo) List(context.Context) ([]model.Category, error) { return nil, nil }

type fakeAuditLogRepo struct {
	entries []*model.AuditLog
}

func (r *fakeAuditLogRepo) Create(_ context.Context, entry *model.AuditLog) error {
	r.entries = append(r.entries, entry)
	return nil
}

// newTestProductService wires ProductService with fakes so audit-log
// behavior can be verified without a real database.
func newTestProductService() (*ProductService, *fakeAuditLogRepo) {
	audit := &fakeAuditLogRepo{}
	svc := NewProductService(newFakeProductRepo(), fakeCategoryRepo{}, audit)
	return svc, audit
}

func TestProductService_Create_RecordsAuditLog(t *testing.T) {
	svc, audit := newTestProductService()

	product, err := svc.Create(context.Background(), 42, model.CreateProductRequest{
		Name: "Kopi Susu", Price: 18000, Stock: 10, CategoryID: 1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if len(audit.entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.ActorID != 42 {
		t.Errorf("ActorID = %d, want 42", entry.ActorID)
	}
	if entry.Action != "product.create" {
		t.Errorf("Action = %q, want product.create", entry.Action)
	}
	if entry.Resource != "product" || entry.ResourceID != product.ID {
		t.Errorf("Resource/ResourceID = %q/%d, want product/%d", entry.Resource, entry.ResourceID, product.ID)
	}
}

func TestProductService_Update_RecordsChangedFieldsInAuditLog(t *testing.T) {
	svc, audit := newTestProductService()

	product, err := svc.Create(context.Background(), 1, model.CreateProductRequest{
		Name: "Kopi Susu", Price: 18000, Stock: 10, CategoryID: 1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	audit.entries = nil // isolate Update's own audit entry

	newStock := 5
	_, err = svc.Update(context.Background(), 42, product.ID, model.UpdateProductRequest{Stock: &newStock})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if len(audit.entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.Action != "product.update" {
		t.Errorf("Action = %q, want product.update", entry.Action)
	}
	if entry.Metadata == "" {
		t.Error("Metadata is empty, want changed fields recorded")
	}
}

// TestProductService_Update_AllowsClearingImageURL is a regression test:
// UpdateProductRequest.ImageURL is a non-nil *string pointing at "" when a
// client explicitly clears the image (found via the admin panel's browser
// verification, see .claude/CLAUDE.md). The validator's `omitempty` tag
// only skips a *nil* pointer, not a non-nil pointer to an empty string, so
// the old `binding:"omitempty,url"` tag rejected this with a 400 — fixed by
// dropping the tag and checking manually in Update, only when non-empty.
func TestProductService_Update_AllowsClearingImageURL(t *testing.T) {
	svc, _ := newTestProductService()

	product, err := svc.Create(context.Background(), 1, model.CreateProductRequest{
		Name: "Kopi Susu", Price: 18000, Stock: 10, CategoryID: 1, ImageURL: "https://example.com/old.jpg",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	emptyImageURL := ""
	updated, err := svc.Update(context.Background(), 1, product.ID, model.UpdateProductRequest{ImageURL: &emptyImageURL})
	if err != nil {
		t.Fatalf("Update with empty ImageURL returned error: %v, want success", err)
	}
	if updated.ImageURL != "" {
		t.Errorf("ImageURL = %q, want cleared to empty", updated.ImageURL)
	}
}

func TestProductService_Update_RejectsMalformedImageURL(t *testing.T) {
	svc, _ := newTestProductService()

	product, err := svc.Create(context.Background(), 1, model.CreateProductRequest{
		Name: "Kopi Susu", Price: 18000, Stock: 10, CategoryID: 1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	notAURL := "not-a-url"
	_, err = svc.Update(context.Background(), 1, product.ID, model.UpdateProductRequest{ImageURL: &notAURL})
	if err == nil {
		t.Fatal("Update with malformed ImageURL returned no error, want validation error")
	}
}

func TestProductService_Delete_RecordsAuditLogAndRemovesProduct(t *testing.T) {
	svc, audit := newTestProductService()

	product, err := svc.Create(context.Background(), 1, model.CreateProductRequest{
		Name: "Kopi Susu", Price: 18000, Stock: 10, CategoryID: 1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	audit.entries = nil // isolate Delete's own audit entry

	if err := svc.Delete(context.Background(), 42, product.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	found, err := svc.productRepo.FindByID(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("FindByID after delete returned error: %v", err)
	}
	if found != nil {
		t.Error("product still findable after Delete")
	}

	if len(audit.entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.Action != "product.delete" || entry.ActorID != 42 || entry.ResourceID != product.ID {
		t.Errorf("audit entry = %+v, want product.delete by actor 42 for resource %d", entry, product.ID)
	}
}

func TestProductService_Delete_NotFound(t *testing.T) {
	svc, _ := newTestProductService()

	err := svc.Delete(context.Background(), 42, 999)
	if err == nil {
		t.Fatal("Delete of nonexistent product returned no error, want NotFound")
	}
}
