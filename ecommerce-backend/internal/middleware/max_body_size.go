package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
)

// maxRequestBodyBytes caps every request body — this API is JSON-only (no
// file uploads), so 1MB is already generous. Cheap insurance against a
// client, malicious or just buggy, sending an oversized body to waste
// memory/bandwidth (see SECURITY_AND_NFR.md baseline).
const maxRequestBodyBytes = 1 << 20 // 1MB

// MaxBodySize rejects a request whose declared Content-Length already
// exceeds the limit before touching the body at all — cheapest possible
// rejection, no read required. It also wraps the body reader with
// http.MaxBytesReader as defense-in-depth for requests that omit
// Content-Length (e.g. chunked transfer encoding), which fails the read
// inside the handler's own binding instead.
func MaxBodySize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxRequestBodyBytes {
			_ = c.Error(apperror.PayloadTooLarge(
				fmt.Sprintf("Request body melebihi batas maksimum %d bytes", maxRequestBodyBytes),
			))
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
		c.Next()
	}
}
