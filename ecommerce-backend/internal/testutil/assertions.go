package testutil

import (
	"testing"

	"ecommerce-backend/internal/apperror"
)

// AssertAppError fails the test unless err is an *apperror.AppError with
// exactly wantCode and wantStatus — asserting both catches two different
// bugs: the wrong Code chosen at the call site, and Code.HTTPStatus()'s
// mapping itself drifting out of sync (a separate method, not derived at
// the call site). Prefer this over asserting err.Error() against a literal
// string, which breaks the moment wording changes for a UX reason
// unrelated to correctness.
func AssertAppError(t *testing.T, err error, wantCode apperror.Code, wantStatus int) {
	t.Helper()

	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("error = %v (%T), want *apperror.AppError with code %s", err, err, wantCode)
	}
	if appErr.Code != wantCode {
		t.Errorf("Code = %s, want %s", appErr.Code, wantCode)
	}
	if appErr.HTTPStatus() != wantStatus {
		t.Errorf("HTTPStatus() = %d, want %d", appErr.HTTPStatus(), wantStatus)
	}
}
