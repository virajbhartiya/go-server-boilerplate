package validator

import (
	"go-server-boilerplate/internal/pkg/errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator is a wrapper around validator.Validate
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator
func New() *Validator {
	v := validator.New()

	// Register custom validations if needed
	// Example: v.RegisterValidation("custom_tag", customValidationFunc)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		// Extract validation errors
		var details []string
		for _, err := range err.(validator.ValidationErrors) {
			details = append(details, formatValidationError(err))
		}
		return errors.BadRequest("Validation failed: " + strings.Join(details, "; "))
	}
	return nil
}

// formatValidationError formats a validation error
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + param
	case "max":
		return field + " must be at most " + param
	case "email":
		return field + " must be a valid email address"
	default:
		return field + " failed " + tag + " validation"
	}
}
