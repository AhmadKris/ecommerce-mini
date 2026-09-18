package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/testutil"
)

// mockCategoryRepository is a testify/mock-based fake — CategoryRepository
// has 5 methods (>3), so per this project's testing convention it's
// mocked with testify/mock rather than hand-rolled (see
// .claude/CLAUDE.md §Testing).
type mockCategoryRepository struct {
	mock.Mock
}

func (m *mockCategoryRepository) Create(ctx context.Context, category *model.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCategoryRepository) Update(ctx context.Context, category *model.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCategoryRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCategoryRepository) FindByID(ctx context.Context, id uint) (*model.Category, error) {
	args := m.Called(ctx, id)
	category, _ := args.Get(0).(*model.Category)
	return category, args.Error(1)
}

func (m *mockCategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	args := m.Called(ctx)
	categories, _ := args.Get(0).([]model.Category)
	return categories, args.Error(1)
}

func TestCategoryService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		req        model.CreateCategoryRequest
		setupMock  func(*mockCategoryRepository)
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{
			name: "success creates category and records audit log",
			req:  model.CreateCategoryRequest{Name: "Minuman"},
			setupMock: func(m *mockCategoryRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *model.Category) bool { return c.Slug == "minuman" })).
					Run(func(args mock.Arguments) { args.Get(1).(*model.Category).ID = 1 }).
					Return(nil).Once()
			},
		},
		{
			name: "slug collision retries with numeric suffix",
			req:  model.CreateCategoryRequest{Name: "Minuman"},
			setupMock: func(m *mockCategoryRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *model.Category) bool { return c.Slug == "minuman" })).
					Return(repository.ErrCategorySlugTaken).Once()
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *model.Category) bool { return c.Slug == "minuman-2" })).
					Run(func(args mock.Arguments) { args.Get(1).(*model.Category).ID = 2 }).
					Return(nil).Once()
			},
		},
		{
			name: "unexpected repository error becomes Internal",
			req:  model.CreateCategoryRequest{Name: "Minuman"},
			setupMock: func(m *mockCategoryRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(gorm.ErrInvalidDB).Once()
			},
			wantErr: true, wantCode: apperror.CodeInternal, wantStatus: 500,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			categoryRepo := new(mockCategoryRepository)
			tc.setupMock(categoryRepo)
			audit := &fakeAuditLogRepo{}
			svc := NewCategoryService(categoryRepo, audit)

			category, err := svc.Create(context.Background(), 42, tc.req)

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if category == nil {
					t.Fatal("expected a category, got nil")
				}
				if len(audit.entries) != 1 || audit.entries[0].Action != "category.create" {
					t.Errorf("expected 1 category.create audit entry, got %+v", audit.entries)
				}
			}
			categoryRepo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMock  func(*mockCategoryRepository)
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{
			name: "success renames and regenerates slug",
			setupMock: func(m *mockCategoryRepository) {
				m.On("FindByID", mock.Anything, uint(1)).Return(testutil.NewCategory(), nil).Once()
				m.On("Update", mock.Anything, mock.MatchedBy(func(c *model.Category) bool {
					return c.Name == "Snack" && c.Slug == "snack"
				})).Return(nil).Once()
			},
		},
		{
			name: "category not found",
			setupMock: func(m *mockCategoryRepository) {
				m.On("FindByID", mock.Anything, uint(1)).Return(nil, nil).Once()
			},
			wantErr: true, wantCode: apperror.CodeNotFound, wantStatus: 404,
		},
		{
			name: "new slug collides with another category",
			setupMock: func(m *mockCategoryRepository) {
				m.On("FindByID", mock.Anything, uint(1)).Return(testutil.NewCategory(), nil).Once()
				m.On("Update", mock.Anything, mock.Anything).Return(repository.ErrCategorySlugTaken).Once()
			},
			wantErr: true, wantCode: apperror.CodeDuplicateEntry, wantStatus: 409,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			categoryRepo := new(mockCategoryRepository)
			tc.setupMock(categoryRepo)
			audit := &fakeAuditLogRepo{}
			svc := NewCategoryService(categoryRepo, audit)

			_, err := svc.Update(context.Background(), 42, 1, model.UpdateCategoryRequest{Name: "Snack"})

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			categoryRepo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		repoErr    error
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{name: "success", repoErr: nil},
		{
			name:    "category still referenced by products",
			repoErr: repository.ErrCategoryInUse, wantErr: true,
			wantCode: apperror.CodeConflict, wantStatus: 409,
		},
		{
			name:    "category not found",
			repoErr: gorm.ErrRecordNotFound, wantErr: true,
			wantCode: apperror.CodeNotFound, wantStatus: 404,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			categoryRepo := new(mockCategoryRepository)
			categoryRepo.On("Delete", mock.Anything, uint(1)).Return(tc.repoErr).Once()
			audit := &fakeAuditLogRepo{}
			svc := NewCategoryService(categoryRepo, audit)

			err := svc.Delete(context.Background(), 42, 1)

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			categoryRepo.AssertExpectations(t)
		})
	}
}
