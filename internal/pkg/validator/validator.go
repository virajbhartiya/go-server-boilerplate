package validator

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"go-server-boilerplate/internal/pkg/errors"
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

// ValidateJSON binds and validates JSON data from a gin context
func (v *Validator) ValidateJSON(c *gin.Context, i interface{}) error {
	if err := c.ShouldBindJSON(i); err != nil {
		return errors.BadRequest("Invalid JSON: " + err.Error())
	}
	return v.Validate(i)
}

// ValidateQuery binds and validates query parameters from a gin context
func (v *Validator) ValidateQuery(c *gin.Context, i interface{}) error {
	if err := c.ShouldBindQuery(i); err != nil {
		return errors.BadRequest("Invalid query parameters: " + err.Error())
	}
	return v.Validate(i)
}

// ValidateURI binds and validates URI parameters from a gin context
func (v *Validator) ValidateURI(c *gin.Context, i interface{}) error {
	if err := c.ShouldBindUri(i); err != nil {
		return errors.BadRequest("Invalid URI parameters: " + err.Error())
	}
	return v.Validate(i)
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
