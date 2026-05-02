package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_AddError(t *testing.T) {
	v := New()
	v.AddError("field1", "error message")

	assert.True(t, v.HasErrors())
	assert.Equal(t, 1, len(v.Errors()))
	assert.Equal(t, "field1", v.Errors()[0].Field)
	assert.Equal(t, "error message", v.Errors()[0].Message)
}

func TestValidator_AddMultipleErrors(t *testing.T) {
	v := New()
	v.AddError("field1", "error1")
	v.AddError("field2", "error2")

	assert.True(t, v.HasErrors())
	assert.Equal(t, 2, len(v.Errors()))
}

func TestValidator_NoErrors(t *testing.T) {
	v := New()
	assert.False(t, v.HasErrors())
	assert.Equal(t, 0, len(v.Errors()))
	assert.Equal(t, 0, len(v.ErrorMap()))
}

func TestValidator_ErrorMap(t *testing.T) {
	v := New()
	v.AddError("email", "Invalid email")
	v.AddError("name", "Name is required")

	errMap := v.ErrorMap()

	assert.Equal(t, 2, len(errMap))
	assert.Equal(t, "Invalid email", errMap["email"])
	assert.Equal(t, "Name is required", errMap["name"])
}

func TestValidator_Required_Empty(t *testing.T) {
	v := New()
	v.Required("email", "")

	assert.True(t, v.HasErrors())
	assert.Contains(t, v.ErrorMap()["email"], "required")
}

func TestValidator_Required_Whitespace(t *testing.T) {
	v := New()
	v.Required("email", "   ")

	assert.True(t, v.HasErrors())
	assert.Contains(t, v.ErrorMap()["email"], "required")
}

func TestValidator_Required_Valid(t *testing.T) {
	v := New()
	v.Required("email", "test@example.com")

	assert.False(t, v.HasErrors())
}

func TestValidator_MinLength_TooShort(t *testing.T) {
	v := New()
	v.MinLength("password", "abc", 5)

	assert.True(t, v.HasErrors())
	assert.Contains(t, v.ErrorMap()["password"], "at least 5 characters")
}

func TestValidator_MinLength_Valid(t *testing.T) {
	v := New()
	v.MinLength("password", "abcdef", 5)

	assert.False(t, v.HasErrors())
}

func TestValidator_MaxLength_TooLong(t *testing.T) {
	v := New()
	v.MaxLength("name", "This is a very long string that exceeds the limit", 10)

	assert.True(t, v.HasErrors())
	assert.Contains(t, v.ErrorMap()["name"], "at most 10 characters")
}

func TestValidator_MaxLength_Valid(t *testing.T) {
	v := New()
	v.MaxLength("name", "Short", 10)

	assert.False(t, v.HasErrors())
}

func TestValidator_Email_Valid(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"123@example.com",
	}

	for _, email := range validEmails {
		v := New()
		v.Email("email", email)
		assert.False(t, v.HasErrors(), "Email should be valid: %s", email)
	}
}

func TestValidator_Email_Invalid(t *testing.T) {
	invalidEmails := []string{
		"invalid",
		"@example.com",
		"user@",
		"user @example.com",
	}

	for _, email := range invalidEmails {
		v := New()
		v.Email("email", email)
		assert.True(t, v.HasErrors(), "Email should be invalid: %s", email)
		assert.Contains(t, v.ErrorMap()["email"], "Invalid email format")
	}
}

func TestValidator_UUID_Valid(t *testing.T) {
	validUUIDs := []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for _, uuid := range validUUIDs {
		v := New()
		v.UUID("id", uuid)
		assert.False(t, v.HasErrors(), "UUID should be valid: %s", uuid)
	}
}

func TestValidator_UUID_Invalid(t *testing.T) {
	invalidUUIDs := []string{
		"not-a-uuid",
		"00000000-0000-0000-0000",
		"ffffffff-ffff-ffff-ffff-fffffffffffff",
		"12345",
	}

	for _, uuid := range invalidUUIDs {
		v := New()
		v.UUID("id", uuid)
		assert.True(t, v.HasErrors(), "UUID should be invalid: %s", uuid)
		assert.Contains(t, v.ErrorMap()["id"], "Invalid UUID format")
	}
}

func TestValidator_Cron_Valid(t *testing.T) {
_validCronExpressions := []string{
		"* * * * *",
		"0 0 * * *",
		"*/5 * * * *",
		"0 9-17 * * 1-5",
		"@hourly",
		"@daily",
		"@weekly",
		"@monthly",
		"@yearly",
		"@every 1h",
		"@every 30m",
	}

	for _, cron := range _validCronExpressions {
		v := New()
		v.Cron("schedule", cron)
		assert.False(t, v.HasErrors(), "Cron should be valid: %s", cron)
	}
}

func TestValidator_Cron_Invalid(t *testing.T) {
	invalidCrons := []string{
		"invalid",
		"* * *",
		"abc * * * *",
	}

	for _, cron := range invalidCrons {
		v := New()
		v.Cron("schedule", cron)
		assert.True(t, v.HasErrors(), "Cron should be invalid: %s", cron)
		assert.Contains(t, v.ErrorMap()["schedule"], "Invalid cron expression")
	}
}

func TestValidator_OneOf_Valid(t *testing.T) {
	v := New()
	v.OneOf("status", "active", []string{"active", "inactive", "pending"})

	assert.False(t, v.HasErrors())
}

func TestValidator_OneOf_Invalid(t *testing.T) {
	v := New()
	v.OneOf("status", "unknown", []string{"active", "inactive", "pending"})

	assert.True(t, v.HasErrors())
	assert.Contains(t, v.ErrorMap()["status"], "must be one of")
}

func TestValidator_Sanitize(t *testing.T) {
	v := New()

	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"<script>alert('xss')</script>", "alert('xss')"},
		{"<p>Hello</p>", "Hello"},
		{"  <div>Test</div>  ", "Test"},
	}

	for _, tt := range tests {
		result := v.Sanitize(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestValidator_SanitizeString(t *testing.T) {
	v := New()

	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"test\x00null", "testnull"},
		{"hello\nworld", "hello\nworld"},
		{"hello\tworld", "hello\tworld"},
	}

	for _, tt := range tests {
		result := v.SanitizeString(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestValidateWorkflowName_Empty(t *testing.T) {
	errors := ValidateWorkflowName("")

	assert.Contains(t, errors, "name is required")
}

func TestValidateWorkflowName_TooShort(t *testing.T) {
	errors := ValidateWorkflowName("ab")

	assert.Contains(t, errors, "name must be at least 3 characters")
}

func TestValidateWorkflowName_TooLong(t *testing.T) {
	longName := string(make([]byte, 256)) // 256 characters
	errors := ValidateWorkflowName(longName)

	assert.Contains(t, errors, "name must be at most 255 characters")
}

func TestValidateWorkflowName_SQLInjection(t *testing.T) {
	sqlInjectionAttempts := []string{
		"test';--",
		"test'; DROP TABLE users;--",
	}

	for _, attempt := range sqlInjectionAttempts {
		errors := ValidateWorkflowName(attempt)
		assert.NotEmpty(t, errors, "Should detect SQL injection in: %s", attempt)

		// Check if any error message contains "invalid characters"
		found := false
		for _, err := range errors {
			if strings.Contains(err, "invalid characters") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find 'invalid characters' in error messages for: %s", attempt)
	}
}

func TestValidateWorkflowName_Valid(t *testing.T) {
	validNames := []string{
		"My Workflow",
		"Test-Workflow_123",
		"Workflow 2024",
		"api Integration",
	}

	for _, name := range validNames {
		errors := ValidateWorkflowName(name)
		assert.Empty(t, errors, "Name should be valid: %s", name)
	}
}

func TestValidateDescription_TooLong(t *testing.T) {
	longDesc := string(make([]byte, 5001)) // 5001 characters
	errors := ValidateDescription(longDesc)

	assert.Contains(t, errors, "description must be at most 5000 characters")
}

func TestValidateDescription_Valid(t *testing.T) {
	errors := ValidateDescription("This is a valid description")

	assert.Empty(t, errors)
}

func TestValidateDescription_Empty(t *testing.T) {
	errors := ValidateDescription("")

	assert.Empty(t, errors) // Description is optional
}

func TestValidateStatus_Valid(t *testing.T) {
	validStatuses := []string{"pending", "running", "success", "failed", "cancelled"}

	for _, status := range validStatuses {
		assert.True(t, ValidateStatus(status), "Status should be valid: %s", status)
	}
}

func TestValidateStatus_Invalid(t *testing.T) {
	assert.False(t, ValidateStatus("unknown"))
	assert.False(t, ValidateStatus("PENDING")) // Case sensitive
	assert.False(t, ValidateStatus(""))
}

func TestValidateRole_Valid(t *testing.T) {
	validRoles := []string{"admin", "editor", "viewer"}

	for _, role := range validRoles {
		assert.True(t, ValidateRole(role), "Role should be valid: %s", role)
	}
}

func TestValidateRole_Invalid(t *testing.T) {
	assert.False(t, ValidateRole("superadmin"))
	assert.False(t, ValidateRole("ADMIN")) // Case sensitive
	assert.False(t, ValidateRole(""))
}

func TestValidateOrderBy_Valid(t *testing.T) {
	validFields := []string{"created_at", "updated_at", "name", "status"}

	for _, field := range validFields {
		assert.True(t, ValidateOrderBy(field), "Field should be valid: %s", field)
	}
}

func TestValidateOrderBy_Invalid(t *testing.T) {
	assert.False(t, ValidateOrderBy("invalid_field"))
	assert.False(t, ValidateOrderBy("id"))
	assert.False(t, ValidateOrderBy(""))
}

func TestValidateOrderDir_Valid(t *testing.T) {
	assert.True(t, ValidateOrderDir("asc"))
	assert.True(t, ValidateOrderDir("desc"))
}

func TestValidateOrderDir_Invalid(t *testing.T) {
	assert.False(t, ValidateOrderDir("ASC")) // Case sensitive
	assert.False(t, ValidateOrderDir("ascending"))
	assert.False(t, ValidateOrderDir(""))
}

func TestValidatePage_Valid(t *testing.T) {
	assert.True(t, ValidatePage(1))
	assert.True(t, ValidatePage(10))
	assert.True(t, ValidatePage(100))
}

func TestValidatePage_Invalid(t *testing.T) {
	assert.False(t, ValidatePage(0))
	assert.False(t, ValidatePage(-1))
}

func TestValidatePerPage_Valid(t *testing.T) {
	assert.True(t, ValidatePerPage(1))
	assert.True(t, ValidatePerPage(10))
	assert.True(t, ValidatePerPage(50))
	assert.True(t, ValidatePerPage(100))
}

func TestValidatePerPage_Invalid(t *testing.T) {
	assert.False(t, ValidatePerPage(0))
	assert.False(t, ValidatePerPage(-1))
	assert.False(t, ValidatePerPage(101))
	assert.False(t, ValidatePerPage(1000))
}
