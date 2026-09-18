package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/logger"
)

// ErrorHandler is the single place that turns an error attached via
// c.Error(err) into a JSON response. Handlers never write error responses
// themselves — they just do `_ = c.Error(err); return`.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err

		appErr, ok := apperror.As(err)
		if !ok {
			logger.FromContext(c.Request.Context()).Error("unhandled error", slog.Any("err", err))
			appErr = apperror.Internal(err)
		}

		if appErr.Code == apperror.CodeInternal {
			logger.FromContext(c.Request.Context()).Error("internal error", slog.Any("err", appErr.Unwrap()))
		}

		c.JSON(appErr.HTTPStatus(), gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"details": appErr.Errors,
			},
		})
	}
}
