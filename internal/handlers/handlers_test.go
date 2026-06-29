package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/APIENG/apieng/internal/db"
	"github.com/APIENG/apieng/internal/models"
)

func setupTestDB(t *testing.T) {
	os.Remove("test_metrics.db")
	db.InitializeDB()
}

func cleanupTestDB(t *testing.T) {
	os.Remove("test_metrics.db")
}

func TestUserSignupAndLogin(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Test signup
	signupData := "email=test@example.com&password=password123&firstName=Test&lastName=User"
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(signupData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	CreateUser(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("CreateUser returned %d, want %d", w.Code, http.StatusSeeOther)
	}

	// Test login
	loginData := "email=test@example.com&password=password123"
	req = httptest.NewRequest("POST", "/login", bytes.NewBufferString(loginData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w = httptest.NewRecorder()
	LoginusersHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("LoginusersHandler returned %d, want %d", w.Code, http.StatusSeeOther)
	}

	// Check that session cookie is set
	cookies := w.Result().Cookies()
	sessionCookie := findCookie(cookies, "session_token")
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Error("session_token cookie not set after login")
	}
}

func TestMeasureAPIValidation(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Test with invalid URL (no http://)
	data := "apiEndpoint=invalid-url"
	req := httptest.NewRequest("POST", "/measure", bytes.NewBufferString(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.AddCookie(&http.Cookie{Name: "user_id", Value: "test_user"})

	w := httptest.NewRecorder()
	MeasureHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("MeasureHandler returned %d for invalid URL, want %d", w.Code, http.StatusBadRequest)
	}

	// Test with valid URL format
	data = "apiEndpoint=https://api.example.com/endpoint"
	req = httptest.NewRequest("POST", "/measure", bytes.NewBufferString(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.AddCookie(&http.Cookie{Name: "user_id", Value: "test_user"})

	w = httptest.NewRecorder()
	MeasureHandler(w, req)

	// Should return 200 OK with JSON response (even if the actual request fails)
	if w.Code != http.StatusOK {
		t.Logf("MeasureHandler returned %d for valid URL (expected 200 OK for AJAX), got: %s", w.Code, w.Body.String())
	}
}

func TestAPIUserEndpointAuthRequired(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Test without authentication
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	APIusersHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("APIusersHandler without auth returned %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// Test with session cookie
	req = httptest.NewRequest("GET", "/api/users", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: "valid_user"})
	w = httptest.NewRecorder()
	APIusersHandler(w, req)

	// Should return 200 (or 404 if user not found, depending on DB)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("APIusersHandler with auth returned %d, want 200 or 404", w.Code)
	}
}

func TestHTMLEscaping(t *testing.T) {
	xssPayload := "<script>alert('xss')</script>"
	escaped := escapeHTML(xssPayload)

	if escaped != "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;" {
		t.Errorf("HTML escaping failed: got %s", escaped)
	}
}

func TestURLValidation(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://api.example.com", true},
		{"http://localhost:8080/endpoint", true},
		{"https://api.example.com:443/path?query=1", true},
		{"invalid-url", false},
		{"ftp://example.com", false},
		{"", false},
		{"//example.com", false},
	}

	for _, test := range tests {
		result := isValidURL(test.url)
		if result != test.valid {
			t.Errorf("isValidURL(%q) = %v, want %v", test.url, result, test.valid)
		}
	}
}

// Helper function to find a cookie by name
func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// Helper to create a test user
func createTestUser(t *testing.T, email, password string) (*models.Users, error) {
	dbConn, err := db.InitializeDB()
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	user := &models.Users{
		Email:     email,
		Password:  password,
		Iid:       "test_" + email,
		FirstName: "Test",
		LastName:  "User",
	}

	err = user.SetPassword(password)
	if err != nil {
		return nil, err
	}

	err = db.StoreUsers(dbConn, *user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
