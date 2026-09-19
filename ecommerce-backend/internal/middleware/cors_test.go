package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestCORS_AllowsCustomHeadersAndMethods is a regression test for a bug
// found via real browser verification (not curl, which never enforces
// CORS): idempotent checkout's Idempotency-Key header and PATCH-based
// endpoints (order/review status updates) were silently rejected by the
// browser's CORS preflight even though curl calls to the same endpoints
// always succeeded.
func TestCORS_AllowsCustomHeadersAndMethods(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(CORS([]string{"http://localhost:3000"}))
	engine.PATCH("/api/admin/reviews/:id/status", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/api/admin/reviews/1/status", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	req.Header.Set("Access-Control-Request-Headers", "Idempotency-Key")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	allowHeaders := recorder.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(allowHeaders, "Idempotency-Key") {
		t.Errorf("Access-Control-Allow-Headers = %q, want it to include Idempotency-Key", allowHeaders)
	}

	allowMethods := recorder.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "PATCH") {
		t.Errorf("Access-Control-Allow-Methods = %q, want it to include PATCH", allowMethods)
	}
}
