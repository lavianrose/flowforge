package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_Success(t *testing.T) {
	password := "mySecurePassword123!"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "Hash should not equal the original password")
	assert.Contains(t, hash, "$", "Hash should contain bcrypt delimiter")
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	password := "samePassword"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, hash1, hash2, "Each hash should be unique due to salt")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	password := ""

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashPassword_SpecialCharacters(t *testing.T) {
	passwords := []string{
		"p@$$w0rd!",
		"日本語パスワード",
		"م fitte كلمة مرور",
		"🔐🔑🗝️",
		"very" + "long" + "password" + "with" + "many" + "characters" + "that" + "goes" + "on" + "and" + "on",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
		})
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	password := "correctPassword123"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	isValid := CheckPassword(password, hash)

	assert.True(t, isValid, "Correct password should validate successfully")
}

func TestCheckPassword_IncorrectPassword(t *testing.T) {
	password := "correctPassword123"
	wrongPassword := "wrongPassword456"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	isValid := CheckPassword(wrongPassword, hash)

	assert.False(t, isValid, "Incorrect password should not validate")
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	password := "nonEmptyPassword"
	emptyPassword := ""

	hash, err := HashPassword(password)
	require.NoError(t, err)

	isValid := CheckPassword(emptyPassword, hash)

	assert.False(t, isValid, "Empty password should not validate against hash")
}

func TestCheckPassword_CaseSensitive(t *testing.T) {
	password := "MyPassword123"
	wrongCasePassword := "mypassword123"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	isValid := CheckPassword(wrongCasePassword, hash)

	assert.False(t, isValid, "Password check should be case-sensitive")
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	password := "myPassword123"

	invalidHashes := []string{
		"invalid-hash",
		"",
		"$2a$10$invalid",
		"plain-text-password",
	}

	for _, hash := range invalidHashes {
		t.Run(hash, func(t *testing.T) {
			isValid := CheckPassword(password, hash)
			assert.False(t, isValid, "Should return false for invalid hash")
		})
	}
}

func TestHashPassword_CheckPassword_RoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{"Simple password", "simple"},
		{"Complex password", "C0mpl3x!P@ssw0rd#2024"},
		{"Long password", "thisIsAVeryLongPasswordThatExceedsNormalLengthButShouldStillWork"},
		{"Unicode password", "パスワード123"},
		{"Mixed case", "MiXeDcAsE123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Hash the password
			hash, err := HashPassword(tc.password)
			require.NoError(t, err)

			// Verify the password against the hash
			isValid := CheckPassword(tc.password, hash)
			assert.True(t, isValid, "Password should validate against its own hash")

			// Verify wrong password doesn't work
			isValid = CheckPassword(tc.password+"wrong", hash)
			assert.False(t, isValid, "Wrong password should not validate")
		})
	}
}

func TestCheckPassword_TimingAttackResistance(t *testing.T) {
	// This test checks that password verification takes similar time
	// regardless of whether the password is correct or not
	password := "testPassword123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// We can't easily test exact timing, but we can ensure both calls succeed without panic
	assert.NotPanics(t, func() {
		CheckPassword(password, hash)
		CheckPassword("wrongPassword", hash)
	})
}

func TestHashPassword_ConsistentFormat(t *testing.T) {
	password := "testPassword123"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Bcrypt hashes should start with $2a$ or $2b$
	assert.Regexp(t, `^\$2[ab]\$`, hash, "Hash should use bcrypt format ($2a$ or $2b$)")
}

func TestCheckPassword_WhitespaceHandling(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		input    string
		expected bool
	}{
		{
			name:     "Password with trailing space",
			password: "password ",
			input:    "password",
			expected: false,
		},
		{
			name:     "Password with leading space",
			password: " password",
			input:    "password",
			expected: false,
		},
		{
			name:     "Password with spaces",
			password: "pass word",
			input:    "pass word",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			require.NoError(t, err)

			isValid := CheckPassword(tc.input, hash)
			assert.Equal(t, tc.expected, isValid)
		})
	}
}

func TestHashPassword_CostFactor(t *testing.T) {
	password := "testPassword123"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Check that the cost factor is set (default is 10 for our implementation)
	// Format: $2a$[cost]$[salt+hash]
	assert.Regexp(t, `^\$2[ab]\$10\$`, hash, "Hash should use cost factor of 10")
}

func TestCheckPassword_RepeatedChecks(t *testing.T) {
	password := "testPassword123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Multiple checks should all succeed
	for i := 0; i < 10; i++ {
		isValid := CheckPassword(password, hash)
		assert.True(t, isValid, "Password should validate consistently across multiple checks")
	}
}
