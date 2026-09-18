// Package handler is the HTTP layer: bind the request, call one service
// method, forward the result. Business conditions are never branched on
// here — that belongs in the service layer.
package handler

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// bindingErrors turns a gin ShouldBindJSON error into per-field messages
// suitable for apperror.Validation's Errors slice. Non-validator errors
// (malformed JSON, wrong type) fall back to the raw error message.
func bindingErrors(err error) []string {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		messages := make([]string, 0, len(validationErrs))
		for _, fieldErr := range validationErrs {
			messages = append(messages, fmt.Sprintf("%s: %s", fieldErr.Field(), fieldErr.Tag()))
		}
		return messages
	}
	return []string{err.Error()}
}
