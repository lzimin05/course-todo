package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/config"
)

func TestNewCookieProvider(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *config.Config
		panics bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				JWTConfig: &config.JWTConfig{
					TokenLifeSpan: 2 * time.Hour,
				},
			},
			panics: false,
		},
		{
			name:   "nil config",
			cfg:    nil,
			panics: false,
		},
		{
			name: "config with nil JWTConfig",
			cfg: &config.Config{
				JWTConfig: nil,
			},
			panics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCookieProvider(tt.cfg)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.cfg, provider.cfg)
		})
	}
}

func TestCookieProvider_Set(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		token       string
		cookieName  string
		expectedAge time.Duration
	}{
		{
			name: "set with custom lifespan",
			cfg: &config.Config{
				JWTConfig: &config.JWTConfig{
					TokenLifeSpan: 2 * time.Hour,
				},
			},
			token:       "test-token-123",
			cookieName:  "auth_token",
			expectedAge: 2 * time.Hour,
		},
		{
			name:        "set with default lifespan (nil config)",
			cfg:         nil,
			token:       "test-token-456",
			cookieName:  "session_token",
			expectedAge: 24 * time.Hour,
		},
		{
			name: "set with nil JWT config",
			cfg: &config.Config{
				JWTConfig: nil,
			},
			token:       "test-token-789",
			cookieName:  "access_token",
			expectedAge: 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCookieProvider(tt.cfg)
			w := httptest.NewRecorder()

			startTime := time.Now().UTC()
			provider.Set(w, tt.token, tt.cookieName)

			cookies := w.Header().Get("Set-Cookie")
			assert.Contains(t, cookies, tt.cookieName+"="+tt.token)
			assert.Contains(t, cookies, "Path=/")
			assert.Contains(t, cookies, "HttpOnly")
			assert.Contains(t, cookies, "SameSite=Strict")

			// Parse the response to get the actual cookie
			resp := &http.Response{Header: w.Header()}
			parsedCookies := resp.Cookies()

			var targetCookie *http.Cookie
			for _, cookie := range parsedCookies {
				if cookie.Name == tt.cookieName {
					targetCookie = cookie
					break
				}
			}

			assert.NotNil(t, targetCookie)
			assert.Equal(t, tt.token, targetCookie.Value)
			assert.Equal(t, "/", targetCookie.Path)
			assert.True(t, targetCookie.HttpOnly)
			assert.Equal(t, http.SameSiteStrictMode, targetCookie.SameSite)

			// Check expiration time (allow some tolerance for test execution time)
			expectedExpiry := startTime.Add(tt.expectedAge)
			timeDiff := targetCookie.Expires.Sub(expectedExpiry)
			assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute,
				"Cookie expiry should be close to expected time, got difference: %v", timeDiff)
		})
	}
}

func TestCookieProvider_Set_EmptyToken(t *testing.T) {
	provider := NewCookieProvider(&config.Config{})
	w := httptest.NewRecorder()

	provider.Set(w, "", "test_cookie")

	// Should not set any cookie when token is empty
	cookies := w.Header().Get("Set-Cookie")
	assert.Empty(t, cookies)
}

func TestCookieProvider_Unset(t *testing.T) {
	tests := []struct {
		name       string
		cookieName string
	}{
		{
			name:       "unset auth token",
			cookieName: "auth_token",
		},
		{
			name:       "unset session token",
			cookieName: "session_token",
		},
		{
			name:       "unset with special characters",
			cookieName: "test_cookie_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCookieProvider(&config.Config{})
			w := httptest.NewRecorder()

			provider.Unset(w, tt.cookieName)

			cookies := w.Header().Get("Set-Cookie")
			assert.Contains(t, cookies, tt.cookieName+"=")
			assert.Contains(t, cookies, "Path=/")
			assert.Contains(t, cookies, "HttpOnly")
			assert.Contains(t, cookies, "Secure")

			// Parse the response to get the actual cookie
			resp := &http.Response{Header: w.Header()}
			parsedCookies := resp.Cookies()

			var targetCookie *http.Cookie
			for _, cookie := range parsedCookies {
				if cookie.Name == tt.cookieName {
					targetCookie = cookie
					break
				}
			}

			assert.NotNil(t, targetCookie)
			assert.Empty(t, targetCookie.Value)
			assert.Equal(t, "/", targetCookie.Path)
			assert.True(t, targetCookie.HttpOnly)
			assert.True(t, targetCookie.Secure)

			// Check that expiration is in the past
			assert.True(t, targetCookie.Expires.Before(time.Now()),
				"Cookie should be expired to unset it")
		})
	}
}

func TestCookieProvider_SetUnsetFlow(t *testing.T) {
	// Test the complete flow of setting and then unsetting a cookie
	provider := NewCookieProvider(&config.Config{
		JWTConfig: &config.JWTConfig{
			TokenLifeSpan: time.Hour,
		},
	})

	cookieName := "test_flow_cookie"
	token := "test-flow-token-123"

	// First set the cookie
	w1 := httptest.NewRecorder()
	provider.Set(w1, token, cookieName)

	setCookies := w1.Header().Get("Set-Cookie")
	assert.Contains(t, setCookies, cookieName+"="+token)

	// Then unset the cookie
	w2 := httptest.NewRecorder()
	provider.Unset(w2, cookieName)

	unsetCookies := w2.Header().Get("Set-Cookie")
	assert.Contains(t, unsetCookies, cookieName+"=")
	assert.NotContains(t, unsetCookies, token)
}

func TestCookieProvider_MultipleTokenLifespans(t *testing.T) {
	lifespans := []time.Duration{
		time.Minute,
		time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
	}

	for _, lifespan := range lifespans {
		t.Run(lifespan.String(), func(t *testing.T) {
			provider := NewCookieProvider(&config.Config{
				JWTConfig: &config.JWTConfig{
					TokenLifeSpan: lifespan,
				},
			})

			w := httptest.NewRecorder()
			startTime := time.Now().UTC()

			provider.Set(w, "token", "test_cookie")

			resp := &http.Response{Header: w.Header()}
			cookies := resp.Cookies()

			assert.Len(t, cookies, 1)
			cookie := cookies[0]

			expectedExpiry := startTime.Add(lifespan)
			timeDiff := cookie.Expires.Sub(expectedExpiry)
			assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute,
				"Cookie expiry for lifespan %v should be correct, got difference: %v", lifespan, timeDiff)
		})
	}
}
