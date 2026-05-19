package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	app, err := New(Config{
		DatabasePath: filepath.Join(t.TempDir(), "pcclub-test.db"),
		SeedDemoData: true,
		JWTSecret:    "test-secret",
		AdminToken:   "admin-token",
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})
	return app
}

func doJSON(t *testing.T, handler http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var reader bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&reader).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, &reader)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeResponse[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()

	var value T
	if err := json.NewDecoder(response.Body).Decode(&value); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, response.Body.String())
	}
	return value
}

func loginTestClient(t *testing.T, handler http.Handler, phone string) AuthResponse {
	t.Helper()

	response := doJSON(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
		"phone":    phone,
		"password": "tester1",
	}, "")
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, body=%s", response.Code, response.Body.String())
	}
	return decodeResponse[AuthResponse](t, response)
}

func TestRegisterAndLoginReturnTokens(t *testing.T) {
	handler := newTestApp(t).Routes()

	registerResponse := doJSON(t, handler, http.MethodPost, "/api/auth/register", map[string]any{
		"name":     "Mobile User",
		"phone":    "89960000099",
		"email":    "mobile@example.com",
		"password": "tester1",
	}, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body=%s", registerResponse.Code, registerResponse.Body.String())
	}
	registerAuth := decodeResponse[AuthResponse](t, registerResponse)
	if registerAuth.ClientID == 0 || registerAuth.Token == "" || registerAuth.RefreshToken == "" {
		t.Fatalf("register auth response is incomplete: %+v", registerAuth)
	}

	loginResponse := doJSON(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
		"phone":    "89960000099",
		"password": "tester1",
	}, "")
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, body=%s", loginResponse.Code, loginResponse.Body.String())
	}
	loginAuth := decodeResponse[AuthResponse](t, loginResponse)
	if loginAuth.ClientID != registerAuth.ClientID || loginAuth.Token == "" || loginAuth.RefreshToken == "" {
		t.Fatalf("login auth response is incomplete: %+v", loginAuth)
	}
}

func TestRejectsClientIDMismatchForOrders(t *testing.T) {
	handler := newTestApp(t).Routes()
	auth := loginTestClient(t, handler, "89962659984")

	response := doJSON(t, handler, http.MethodPost, "/api/orders", map[string]any{
		"clientId": auth.ClientID + 1,
		"items": []map[string]any{
			{"productId": 1, "quantity": 1},
		},
	}, auth.Token)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body=%s", response.Code, response.Body.String())
	}

	errorBody := decodeResponse[map[string]string](t, response)
	if errorBody["code"] != "CLIENT_ID_MISMATCH" {
		t.Fatalf("code = %q, want CLIENT_ID_MISMATCH", errorBody["code"])
	}
}

func TestBookingConflict(t *testing.T) {
	handler := newTestApp(t).Routes()
	auth := loginTestClient(t, handler, "89962659984")
	startTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339Nano)

	body := map[string]any{
		"clientId":  auth.ClientID,
		"pcName":    "VIP-01",
		"startTime": startTime,
		"duration":  2,
	}

	first := doJSON(t, handler, http.MethodPost, "/api/bookings", body, auth.Token)
	if first.Code != http.StatusCreated {
		t.Fatalf("first booking status = %d, body=%s", first.Code, first.Body.String())
	}

	second := doJSON(t, handler, http.MethodPost, "/api/bookings", body, auth.Token)
	if second.Code != http.StatusConflict {
		t.Fatalf("second booking status = %d, want 409, body=%s", second.Code, second.Body.String())
	}

	errorBody := decodeResponse[map[string]string](t, second)
	if errorBody["code"] != "BOOKING_CONFLICT" {
		t.Fatalf("code = %q, want BOOKING_CONFLICT", errorBody["code"])
	}
}

func TestOrderUsesDatabasePriceAndRejectsMissingProduct(t *testing.T) {
	handler := newTestApp(t).Routes()
	auth := loginTestClient(t, handler, "89962659984")

	orderResponse := doJSON(t, handler, http.MethodPost, "/api/orders", map[string]any{
		"clientId": auth.ClientID,
		"items": []map[string]any{
			{"productId": 1, "quantity": 2},
		},
	}, auth.Token)
	if orderResponse.Code != http.StatusCreated {
		t.Fatalf("order status = %d, body=%s", orderResponse.Code, orderResponse.Body.String())
	}
	order := decodeResponse[Order](t, orderResponse)
	if order.TotalAmount != 180 || order.Items[0].Price != 90 || order.Status != "new" {
		t.Fatalf("unexpected order response: %+v", order)
	}

	missingProductResponse := doJSON(t, handler, http.MethodPost, "/api/orders", map[string]any{
		"clientId": auth.ClientID,
		"items": []map[string]any{
			{"productId": 999, "quantity": 1},
		},
	}, auth.Token)
	if missingProductResponse.Code != http.StatusNotFound {
		t.Fatalf("missing product status = %d, want 404, body=%s", missingProductResponse.Code, missingProductResponse.Body.String())
	}
	errorBody := decodeResponse[map[string]string](t, missingProductResponse)
	if errorBody["code"] != "PRODUCT_NOT_FOUND" {
		t.Fatalf("code = %q, want PRODUCT_NOT_FOUND", errorBody["code"])
	}
}
