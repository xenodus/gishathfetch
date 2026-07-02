package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/adminauth"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/adminlogin"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

type memoryAdminLoginStore struct {
	lockout      adminlogin.Lockout
	attempts     []adminlogin.Attempt
	failures     int
	cleared      int
	recordErr    error
	lockoutErr   error
}

func (m *memoryAdminLoginStore) CheckLockout(
	_ context.Context,
	_, _ string,
	_ time.Time,
) (adminlogin.Lockout, error) {
	if m.lockoutErr != nil {
		return adminlogin.Lockout{}, m.lockoutErr
	}
	return m.lockout, nil
}

func (m *memoryAdminLoginStore) RecordAttempt(
	_ context.Context,
	attempt adminlogin.Attempt,
	_ time.Duration,
) error {
	if m.recordErr != nil {
		return m.recordErr
	}
	m.attempts = append(m.attempts, attempt)
	return nil
}

func (m *memoryAdminLoginStore) RecordFailure(
	_ context.Context,
	_, _ string,
	_ time.Time,
	_ adminlogin.RateLimits,
) error {
	m.failures++
	return nil
}

func (m *memoryAdminLoginStore) ClearFailures(_ context.Context, _, _ string) error {
	m.cleared++
	return nil
}

func withAdminEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.AdminUsernameEnv, "admin")
	t.Setenv(config.AdminPasswordEnv, "secret")
	t.Setenv(config.AdminSessionSecretEnv, "session-secret")
	t.Setenv(config.AdminLoginDynamoDBTableEnv, "admin-login-test")
}

func TestRouteAdmin_DisabledReturnsServiceUnavailable(t *testing.T) {
	loginDelayFunc = func() {}
	t.Cleanup(func() { loginDelayFunc = defaultLoginDelay })

	response, err := adminLogin(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       adminLoginPath,
		Body:       `{"username":"admin","password":"secret"}`,
		Headers:    map[string]string{"origin": "https://gishathfetch.com"},
	}, "https://gishathfetch.com")
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, response.StatusCode)
}

func TestAdminLogin_SuccessSetsSessionCookie(t *testing.T) {
	withAdminEnv(t)
	loginDelayFunc = func() {}
	t.Cleanup(func() { loginDelayFunc = defaultLoginDelay })

	store := &memoryAdminLoginStore{}
	originalFactory := newAdminLoginService
	newAdminLoginService = func(_ context.Context) (*adminlogin.Service, error) {
		return adminlogin.NewService(store, adminlogin.RateLimits{
			MaxFailuresPerIP:   5,
			IPWindow:           15 * time.Minute,
			IPLockout:          15 * time.Minute,
			MaxFailuresPerUser: 10,
			UserWindow:         30 * time.Minute,
			UserLockout:        30 * time.Minute,
		}), nil
	}
	t.Cleanup(func() { newAdminLoginService = originalFactory })

	response, err := adminLogin(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       adminLoginPath,
		Body:       `{"username":"admin","password":"secret"}`,
		Headers: map[string]string{
			"origin": "https://gishathfetch.com",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Identity: events.APIGatewayRequestIdentity{
				SourceIP: "203.0.113.10",
			},
		},
	}, "https://gishathfetch.com")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Contains(t, response.Headers["Set-Cookie"], adminauth.SessionCookieName+"=")
	require.Equal(t, 1, store.cleared)
	require.Len(t, store.attempts, 1)
	require.True(t, store.attempts[0].Success)
}

func TestAdminLogin_InvalidCredentialsIncrementsFailures(t *testing.T) {
	withAdminEnv(t)
	loginDelayFunc = func() {}
	t.Cleanup(func() { loginDelayFunc = defaultLoginDelay })

	store := &memoryAdminLoginStore{}
	originalFactory := newAdminLoginService
	newAdminLoginService = func(_ context.Context) (*adminlogin.Service, error) {
		return adminlogin.NewService(store, adminlogin.RateLimits{
			MaxFailuresPerIP: 5,
			IPWindow:         15 * time.Minute,
			IPLockout:        15 * time.Minute,
		}), nil
	}
	t.Cleanup(func() { newAdminLoginService = originalFactory })

	response, err := adminLogin(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       adminLoginPath,
		Body:       `{"username":"admin","password":"wrong"}`,
		Headers:    map[string]string{"origin": "https://gishathfetch.com"},
	}, "https://gishathfetch.com")
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	require.Equal(t, 1, store.failures)
}

func TestAdminLogin_LockoutReturnsTooManyRequests(t *testing.T) {
	withAdminEnv(t)
	loginDelayFunc = func() {}
	t.Cleanup(func() { loginDelayFunc = defaultLoginDelay })

	store := &memoryAdminLoginStore{
		lockout: adminlogin.Lockout{Locked: true, RetryAfter: time.Minute},
	}
	originalFactory := newAdminLoginService
	newAdminLoginService = func(_ context.Context) (*adminlogin.Service, error) {
		return adminlogin.NewService(store, adminlogin.RateLimits{}), nil
	}
	t.Cleanup(func() { newAdminLoginService = originalFactory })

	response, err := adminLogin(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       adminLoginPath,
		Body:       `{"username":"admin","password":"secret"}`,
		Headers:    map[string]string{"origin": "https://gishathfetch.com"},
	}, "https://gishathfetch.com")
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, response.StatusCode)
	require.Len(t, store.attempts, 1)
	require.True(t, store.attempts[0].Blocked)
}

func TestAdminSession_DisabledReportsNotEnabled(t *testing.T) {
	response, err := adminSession(events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       adminSessionPath,
		Headers:    map[string]string{"origin": "https://gishathfetch.com"},
	}, "https://gishathfetch.com")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	var payload adminSessionResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	require.False(t, payload.Authenticated)
	require.False(t, payload.Enabled)
}

func TestAdminSession_ValidCookieReportsAuthenticated(t *testing.T) {
	withAdminEnv(t)

	token, err := adminauth.IssueSessionToken("session-secret", "admin", time.Hour)
	require.NoError(t, err)

	response, err := adminSession(events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       adminSessionPath,
		Headers: map[string]string{
			"origin":  "https://gishathfetch.com",
			"cookie":  adminauth.SessionCookieName + "=" + token,
		},
	}, "https://gishathfetch.com")
	require.NoError(t, err)

	var payload adminSessionResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	require.True(t, payload.Authenticated)
	require.True(t, payload.Enabled)
}

func TestHandle_RoutesSearchWhenPathIsNotAdmin(t *testing.T) {
	originalSearchFunc := searchFunc
	searchFunc = func(_ context.Context, _ controller.SearchInput) ([]controller.Card, []controller.StoreError, error) {
		return []controller.Card{{Name: "Opt"}}, nil, nil
	}
	t.Cleanup(func() { searchFunc = originalSearchFunc })

	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	payload, err := json.Marshal(events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/",
		QueryStringParameters: map[string]string{
			"s": "Opt",
		},
	})
	require.NoError(t, err)

	response, err := Handle(context.Background(), payload)
	require.NoError(t, err)

	apiResponse, ok := response.(events.APIGatewayProxyResponse)
	require.True(t, ok)
	require.Equal(t, http.StatusOK, apiResponse.StatusCode)
}

func defaultLoginDelay() {
	jitter := time.Duration(time.Now().UnixNano()%200) * time.Millisecond
	time.Sleep(200*time.Millisecond + jitter)
}
