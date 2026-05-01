package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	uuidRegex  = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	cronRegex  = regexp.MustCompile(`^(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*) ?){5,7})$`)
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	errors []ValidationError
}

func New() *Validator {
	return &Validator{errors: make([]ValidationError, 0)}
}

func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *Validator) Errors() []ValidationError {
	return v.errors
}

func (v *Validator) ErrorMap() map[string]string {
	result := make(map[string]string)
	for _, err := range v.errors {
		result[err.Field] = err.Message
	}
	return result
}

func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field))
	}
}

func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters", field, min))
	}
}

func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("%s must be at most %d characters", field, max))
	}
}

func (v *Validator) Email(field, value string) {
	if !emailRegex.MatchString(value) {
		v.AddError(field, "Invalid email format")
	}
}

func (v *Validator) UUID(field, value string) {
	if !uuidRegex.MatchString(value) {
		v.AddError(field, "Invalid UUID format")
	}
}

func (v *Validator) Cron(field, value string) {
	if !cronRegex.MatchString(value) {
		v.AddError(field, "Invalid cron expression")
	}
}

func (v *Validator) OneOf(field, value string, allowed []string) {
	for _, item := range allowed {
		if value == item {
			return
		}
	}
	v.AddError(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")))
}

func (v *Validator) Sanitize(value string) string {
	// Trim whitespace
	value = strings.TrimSpace(value)

	// Remove potential HTML tags (basic)
	value = stripTags(value)

	return value
}

func (v *Validator) SanitizeString(value string) string {
	// Trim and remove null bytes
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\x00", "")

	// Remove control characters except newline, tab, carriage return
	runes := []rune(value)
	var sanitized []rune
	for _, r := range runes {
		if unicode.IsGraphic(r) || r == '\n' || r == '\t' || r == '\r' {
			sanitized = append(sanitized, r)
		}
	}

	return string(sanitized)
}

func stripTags(s string) string {
	inTag := false
	runes := []rune(s)
	var result []rune

	for _, r := range runes {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result = append(result, r)
		}
	}

	return string(result)
}

// ValidateWorkflowName validates workflow name
func ValidateWorkflowName(name string) []string {
	var errors []string

	name = strings.TrimSpace(name)

	if name == "" {
		errors = append(errors, "name is required")
	}

	if len(name) > 255 {
		errors = append(errors, "name must be at most 255 characters")
	}

	if len(name) < 3 {
		errors = append(errors, "name must be at least 3 characters")
	}

	// Check for SQL injection patterns
	sqlPatterns := []string{"';--", "';", "--", "/*", "*/", "xp_", "exec(", "1=1"}
	for _, pattern := range sqlPatterns {
		if strings.Contains(strings.ToLower(name), pattern) {
			errors = append(errors, "name contains invalid characters")
			break
		}
	}

	return errors
}

// ValidateDescription validates workflow description
func ValidateDescription(desc string) []string {
	var errors []string

	if len(desc) > 5000 {
		errors = append(errors, "description must be at most 5000 characters")
	}

	return errors
}

// ValidateStatus validates workflow run status
func ValidateStatus(status string) bool {
	allowed := []string{"pending", "running", "success", "failed", "cancelled"}
	for _, s := range allowed {
		if status == s {
			return true
		}
	}
	return false
}

// ValidateRole validates user role
func ValidateRole(role string) bool {
	allowed := []string{"admin", "editor", "viewer"}
	for _, r := range allowed {
		if role == r {
			return true
		}
	}
	return false
}

// ValidateOrderBy validates order by field
func ValidateOrderBy(field string) bool {
	allowed := []string{"created_at", "updated_at", "name", "status"}
	for _, a := range allowed {
		if field == a {
			return true
		}
	}
	return false
}

// ValidateOrderDir validates order direction
func ValidateOrderDir(dir string) bool {
	return dir == "asc" || dir == "desc"
}

// ValidatePage validates page number
func ValidatePage(page int) bool {
	return page > 0
}

// ValidatePerPage validates per page limit
func ValidatePerPage(perPage int) bool {
	return perPage > 0 && perPage <= 100
}
