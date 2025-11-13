package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dto "github.com/lzimin05/course-todo/internal/transport/dto/auth"
)

func TestValidateRegisterRequest_Success(t *testing.T) {
	validRequest := dto.RegisterRequest{
		Login:    "testuser",
		Username: "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := ValidateRegisterRequest(validRequest)
	assert.NoError(t, err)
}

func TestValidateRegisterRequest_Login_Validation(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		expectedErr string
	}{
		{
			name:        "empty login",
			login:       "",
			expectedErr: "login is required",
		},
		{
			name:        "login with spaces only",
			login:       "   ",
			expectedErr: "login is required",
		},
		{
			name:        "login too short",
			login:       "ab",
			expectedErr: "login must be between 3 and 50 characters",
		},
		{
			name:        "login too long",
			login:       "this_is_a_very_long_login_name_that_exceeds_fifty_characters",
			expectedErr: "login must be between 3 and 50 characters",
		},
		{
			name:        "login with invalid characters",
			login:       "user@domain",
			expectedErr: "login can only contain letters, numbers and underscores",
		},
		{
			name:        "login with spaces",
			login:       "user name",
			expectedErr: "login can only contain letters, numbers and underscores",
		},
		{
			name:        "login with special characters",
			login:       "user-name",
			expectedErr: "login can only contain letters, numbers and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := dto.RegisterRequest{
				Login:    tt.login,
				Username: "Valid Username",
				Email:    "valid@example.com",
				Password: "validpassword123",
			}

			err := ValidateRegisterRequest(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateRegisterRequest_Username_Validation(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectedErr string
	}{
		{
			name:        "empty username",
			username:    "",
			expectedErr: "username is required",
		},
		{
			name:        "username with spaces only",
			username:    "   ",
			expectedErr: "username is required",
		},
		{
			name:        "username too short",
			username:    "a",
			expectedErr: "username must be between 2 and 100 characters",
		},
		{
			name:        "username too long",
			username:    "This is a very long username that definitely exceeds one hundred characters limit and should fail validation test completely",
			expectedErr: "username must be between 2 and 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := dto.RegisterRequest{
				Login:    "validlogin",
				Username: tt.username,
				Email:    "valid@example.com",
				Password: "validpassword123",
			}

			err := ValidateRegisterRequest(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateRegisterRequest_Email_Validation(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectedErr string
	}{
		{
			name:        "empty email",
			email:       "",
			expectedErr: "email is required",
		},
		{
			name:        "email with spaces only",
			email:       "   ",
			expectedErr: "email is required",
		},
		{
			name:        "email too long",
			email:       "very_long_email_address_that_exceeds_the_maximum_allowed_length_of_two_hundred_and_fifty_five_characters_and_should_fail_validation_test_because_it_is_way_too_long_for_a_normal_email_address_to_be_considered_valid_and_this_additional_text_makes_it_even_longer@example.com",
			expectedErr: "email is too long",
		},
		{
			name:        "invalid email format - no @",
			email:       "invalidemail.com",
			expectedErr: "invalid email format",
		},
		{
			name:        "invalid email format - no domain",
			email:       "user@",
			expectedErr: "invalid email format",
		},
		{
			name:        "invalid email format - no TLD",
			email:       "user@domain",
			expectedErr: "invalid email format",
		},
		{
			name:        "invalid email format - multiple @",
			email:       "user@@domain.com",
			expectedErr: "invalid email format",
		},
		{
			name:        "invalid email format - spaces",
			email:       "user @domain.com",
			expectedErr: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := dto.RegisterRequest{
				Login:    "validlogin",
				Username: "Valid Username",
				Email:    tt.email,
				Password: "validpassword123",
			}

			err := ValidateRegisterRequest(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateRegisterRequest_Password_Validation(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectedErr string
	}{
		{
			name:        "empty password",
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "password with spaces only",
			password:    "   ",
			expectedErr: "password is required",
		},
		{
			name:        "password too short",
			password:    "1234567",
			expectedErr: "password must be at least 8 characters long",
		},
		{
			name:        "password too long",
			password:    "this_is_an_extremely_long_password_that_exceeds_seventy_two_characters_limit",
			expectedErr: "password is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := dto.RegisterRequest{
				Login:    "validlogin",
				Username: "Valid Username",
				Email:    "valid@example.com",
				Password: tt.password,
			}

			err := ValidateRegisterRequest(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateRegisterRequest_ValidCases(t *testing.T) {
	validCases := []dto.RegisterRequest{
		{
			Login:    "user123",
			Username: "John Doe",
			Email:    "john.doe@example.com",
			Password: "securepassword123",
		},
		{
			Login:    "test_user",
			Username: "Test User",
			Email:    "test.user+tag@domain.co.uk",
			Password: "MySecurePassword123!",
		},
		{
			Login:    "admin_user",
			Username: "Administrator",
			Email:    "admin@company-domain.org",
			Password: "AdminPassword2023",
		},
	}

	for i, validCase := range validCases {
		t.Run("valid_case_"+string(rune('A'+i)), func(t *testing.T) {
			err := ValidateRegisterRequest(validCase)
			assert.NoError(t, err)
		})
	}
}

func TestValidateRegisterRequest_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		request dto.RegisterRequest
		valid   bool
	}{
		{
			name: "minimum valid lengths",
			request: dto.RegisterRequest{
				Login:    "abc",      // exactly 3 characters
				Username: "AB",       // exactly 2 characters
				Email:    "a@b.co",   // minimum valid email
				Password: "12345678", // exactly 8 characters
			},
			valid: true,
		},
		{
			name: "maximum valid lengths",
			request: dto.RegisterRequest{
				Login:    "abcdefghijklmnopqrstuvwxyz1234567890123456789012",                                          // exactly 50 characters
				Username: "This is a username that is exactly one hundred characters long to test the maximum limit.", // exactly 100 characters
				Email:    "user@example.com",
				Password: "This_password_is_exactly_seventy_two_characters_long_and_should_pass", // exactly 72 characters
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegisterRequest(tt.request)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
