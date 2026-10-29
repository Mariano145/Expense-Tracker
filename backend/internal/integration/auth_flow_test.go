package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Mariano145/spendly/backend/internal/database"
	"github.com/Mariano145/spendly/backend/internal/handler"
	"github.com/Mariano145/spendly/backend/internal/middleware"
	"github.com/Mariano145/spendly/backend/internal/repository"
	"github.com/Mariano145/spendly/backend/internal/router"
	"github.com/Mariano145/spendly/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	server *httptest.Server
	pool   *pgxpool.Pool
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://spendly:spendly@localhost:5432/spendly_test?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := database.NewPostgresDB(ctx, database.DefaultPoolConfig(dbURL))
	require.NoError(t, err, "failed to connect to test database")

	cleanDatabase(t, pool)

	jwtSecret := "test-secret-key-must-be-at-least-32-characters-long"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 168 * time.Hour
	bcryptCost := 4

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewRefreshTokenRepository(pool)

	authSvc := service.NewAuthService(
		userRepo,
		tokenRepo,
		jwtSecret,
		accessExpiry,
		refreshExpiry,
		bcryptCost,
	)

	authHandler := handler.NewAuthHandler(authSvc)
	authMW := middleware.Auth(authSvc)
	corsMW := middleware.CORS("http://localhost:3000")

	h := router.New(router.Deps{
		AuthHandler: authHandler,
		AuthMW:      authMW,
		CORSMW:      corsMW,
	})

	server := httptest.NewServer(h)

	t.Cleanup(func() {
		server.Close()
		pool.Close()
	})

	return &testEnv{server: server, pool: pool}
}

func cleanDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, "DELETE FROM refresh_tokens")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM users")
	require.NoError(t, err)
}

type registerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type authResponse struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func extractRefreshCookie(resp *http.Response) string {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "refresh_token" {
			return cookie.Value
		}
	}
	return ""
}

func TestAuthFlow_RegisterLoginRefreshMeLogout(t *testing.T) {
	env := setupTestEnv(t)
	client := &http.Client{}

	email := "test@example.com"
	password := "securepassword123"

	// Step 1: Register
	regBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	regResp, err := client.Post(env.server.URL+"/api/auth/register", "application/json", bytes.NewReader(regBody))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, regResp.StatusCode)

	var regResult registerResponse
	err = json.NewDecoder(regResp.Body).Decode(&regResult)
	require.NoError(t, err)
	assert.Equal(t, email, regResult.Email)
	assert.NotEmpty(t, regResult.ID)
	regResp.Body.Close()

	// Step 2: Login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	loginResp, err := client.Post(env.server.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginResult authResponse
	err = json.NewDecoder(loginResp.Body).Decode(&loginResult)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResult.AccessToken)
	assert.Equal(t, email, loginResult.User.Email)
	loginResp.Body.Close()

	refreshToken := extractRefreshCookie(loginResp)
	assert.NotEmpty(t, refreshToken, "refresh token cookie should be set")

	// Step 3: Access protected endpoint /api/auth/me
	meReq, err := http.NewRequest("GET", env.server.URL+"/api/auth/me", nil)
	require.NoError(t, err)
	meReq.Header.Set("Authorization", "Bearer "+loginResult.AccessToken)

	meResp, err := client.Do(meReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, meResp.StatusCode)

	var meResult struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	err = json.NewDecoder(meResp.Body).Decode(&meResult)
	require.NoError(t, err)
	assert.Equal(t, email, meResult.Email)
	meResp.Body.Close()

	// Step 4: Refresh token
	refreshReq, err := http.NewRequest("POST", env.server.URL+"/api/auth/refresh", nil)
	require.NoError(t, err)
	refreshReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	refreshResp, err := client.Do(refreshReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, refreshResp.StatusCode)

	var refreshResult refreshResponse
	err = json.NewDecoder(refreshResp.Body).Decode(&refreshResult)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResult.AccessToken)
	refreshResp.Body.Close()

	newRefreshToken := extractRefreshCookie(refreshResp)
	assert.NotEmpty(t, newRefreshToken, "new refresh token cookie should be set")
	assert.NotEqual(t, refreshToken, newRefreshToken, "refresh token should be rotated")

	// Step 5: Access /api/auth/me with new access token
	meReq2, err := http.NewRequest("GET", env.server.URL+"/api/auth/me", nil)
	require.NoError(t, err)
	meReq2.Header.Set("Authorization", "Bearer "+refreshResult.AccessToken)

	meResp2, err := client.Do(meReq2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, meResp2.StatusCode)
	meResp2.Body.Close()

	// Step 6: Logout
	logoutReq, err := http.NewRequest("POST", env.server.URL+"/api/auth/logout", nil)
	require.NoError(t, err)
	logoutReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: newRefreshToken})

	logoutResp, err := client.Do(logoutReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, logoutResp.StatusCode)
	logoutResp.Body.Close()

	// Step 7: Verify refresh token is revoked
	reuseReq, err := http.NewRequest("POST", env.server.URL+"/api/auth/refresh", nil)
	require.NoError(t, err)
	reuseReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: newRefreshToken})

	reuseResp, err := client.Do(reuseReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, reuseResp.StatusCode)
	reuseResp.Body.Close()
}

func TestAuthFlow_TokenTheftDetection(t *testing.T) {
	env := setupTestEnv(t)
	client := &http.Client{}

	email := "theft@example.com"
	password := "securepassword123"

	// Register
	regBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	regResp, err := client.Post(env.server.URL+"/api/auth/register", "application/json", bytes.NewReader(regBody))
	require.NoError(t, err)
	regResp.Body.Close()

	// Login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	loginResp, err := client.Post(env.server.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	loginResp.Body.Close()

	originalRefreshToken := extractRefreshCookie(loginResp)
	require.NotEmpty(t, originalRefreshToken)

	// First refresh (legitimate)
	refreshReq1, err := http.NewRequest("POST", env.server.URL+"/api/auth/refresh", nil)
	require.NoError(t, err)
	refreshReq1.AddCookie(&http.Cookie{Name: "refresh_token", Value: originalRefreshToken})

	refreshResp1, err := client.Do(refreshReq1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, refreshResp1.StatusCode)
	refreshResp1.Body.Close()

	newRefreshToken := extractRefreshCookie(refreshResp1)
	require.NotEmpty(t, newRefreshToken)

	// Attacker reuses the OLD token (theft detection)
	reuseReq, err := http.NewRequest("POST", env.server.URL+"/api/auth/refresh", nil)
	require.NoError(t, err)
	reuseReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: originalRefreshToken})

	reuseResp, err := client.Do(reuseReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, reuseResp.StatusCode)

	var reuseErr errorResponse
	err = json.NewDecoder(reuseResp.Body).Decode(&reuseErr)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", reuseErr.Error.Code)
	reuseResp.Body.Close()

	// Verify the NEW token is also revoked (family revoked)
	familyReq, err := http.NewRequest("POST", env.server.URL+"/api/auth/refresh", nil)
	require.NoError(t, err)
	familyReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: newRefreshToken})

	familyResp, err := client.Do(familyReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, familyResp.StatusCode, "entire family should be revoked after theft detection")
	familyResp.Body.Close()
}

func TestAuthFlow_DuplicateRegistration(t *testing.T) {
	env := setupTestEnv(t)
	client := &http.Client{}

	email := "duplicate@example.com"
	password := "securepassword123"

	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	// First registration
	resp1, err := client.Post(env.server.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)
	resp1.Body.Close()

	// Duplicate registration
	resp2, err := client.Post(env.server.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp2.StatusCode)

	var errResp errorResponse
	err = json.NewDecoder(resp2.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "CONFLICT", errResp.Error.Code)
	resp2.Body.Close()
}

func TestAuthFlow_InvalidCredentials(t *testing.T) {
	env := setupTestEnv(t)
	client := &http.Client{}

	// Register
	regBody, _ := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "correctpassword",
	})
	regResp, err := client.Post(env.server.URL+"/api/auth/register", "application/json", bytes.NewReader(regBody))
	require.NoError(t, err)
	regResp.Body.Close()

	// Login with wrong password
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "wrongpassword",
	})
	loginResp, err := client.Post(env.server.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, loginResp.StatusCode)
	loginResp.Body.Close()
}

func TestAuthFlow_ProtectedEndpointWithoutToken(t *testing.T) {
	env := setupTestEnv(t)
	client := &http.Client{}

	resp, err := client.Get(env.server.URL + "/api/auth/me")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}
