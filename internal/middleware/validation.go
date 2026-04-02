package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Validator handles input validation
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateUUID validates a UUID string
func (v *Validator) ValidateUUID(value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return &ValidationError{Field: "uuid", Message: "invalid UUID format"}
	}
	return nil
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(email) {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}
	return nil
}

// ValidatePassword validates a password
func (v *Validator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	return nil
}

// ValidateRFC validates a Mexican RFC
func (v *Validator) ValidateRFC(rfc string) error {
	if len(rfc) < 12 || len(rfc) > 13 {
		return &ValidationError{Field: "rfc", Message: "RFC must be 12 or 13 characters"}
	}
	return nil
}

// ValidatePhone validates a phone number
func (v *Validator) ValidatePhone(phone string) error {
	if len(phone) < 10 {
		return &ValidationError{Field: "phone", Message: "phone must be at least 10 digits"}
	}
	return nil
}

// ValidateGraphQLRequest validates a GraphQL request body
func (v *Validator) ValidateGraphQLRequest(body []byte) error {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return &ValidationError{Field: "body", Message: "invalid JSON"}
	}
	if req["query"] == nil || req["query"] == "" {
		return &ValidationError{Field: "query", Message: "query is required"}
	}
	return nil
}

// ValidateInputWithTags validates a struct based on validation tags
func (v *Validator) ValidateInputWithTags(data interface{}) []*ValidationError {
	var errors []*ValidationError

	switch d := data.(type) {
	case map[string]interface{}:
		if name, ok := d["name"].(string); ok && name == "" {
			errors = append(errors, &ValidationError{Field: "name", Message: "name is required"})
		}
		if email, ok := d["email"].(string); ok && email != "" {
			if err := v.ValidateEmail(email); err != nil {
				errors = append(errors, err.(*ValidationError))
			}
		}
		if password, ok := d["password"].(string); ok && len(password) < 8 {
			errors = append(errors, &ValidationError{Field: "password", Message: "password must be at least 8 characters"})
		}
	}

	return errors
}

// ValidateRequestBody validates the request body based on content type
func ValidateRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			next.ServeHTTP(w, r)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			http.Error(w, "Content-Type header is required", http.StatusBadRequest)
			return
		}

		if !strings.Contains(contentType, "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	})
}
