package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	ClientID     int64  `json:"clientId"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type authClaims struct {
	ClientID int64  `json:"clientId"`
	Type     string `json:"type"`
	Exp      int64  `json:"exp"`
}

type authError struct {
	status  int
	code    string
	message string
}

func (e authError) Error() string {
	return e.message
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	if !a.authLimiter.Allow(rateLimitKey(r, "register")) {
		writeErrorCode(w, http.StatusTooManyRequests, "RATE_LIMITED", "Слишком много попыток регистрации")
		return
	}

	var request AuthRequest
	if err := decodeJSON(r, &request); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный JSON")
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	request.Phone = normalizePhone(request.Phone)
	request.Email = strings.TrimSpace(request.Email)

	if err := validateClientFields(request.Name, request.Phone, request.Email, false); err != nil {
		writeValidationError(w, err)
		return
	}
	if err := validatePassword(request.Password); err != nil {
		writeValidationError(w, err)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result, err := a.db.Exec(
		`INSERT INTO Clients (Name, Phone, Email, PasswordHash, Balance) VALUES (?, ?, ?, ?, 0)`,
		request.Name, request.Phone, request.Email, string(passwordHash),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeErrorCode(w, http.StatusConflict, "PHONE_EXISTS", "Телефон уже зарегистрирован")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	clientID, _ := result.LastInsertId()
	response, err := a.authResponse(clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if !a.authLimiter.Allow(rateLimitKey(r, "login")) {
		writeErrorCode(w, http.StatusTooManyRequests, "RATE_LIMITED", "Слишком много попыток входа")
		return
	}

	var request AuthRequest
	if err := decodeJSON(r, &request); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный JSON")
		return
	}
	request.Phone = normalizePhone(request.Phone)
	if err := requireText(request.Phone, "phone"); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_PHONE", "Телефон обязателен")
		return
	}
	if err := requireText(request.Password, "password"); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_PASSWORD", "Пароль обязателен")
		return
	}

	var clientID int64
	var passwordHash sql.NullString
	err := a.db.QueryRow(`SELECT Id, PasswordHash FROM Clients WHERE Phone = ?`, request.Phone).Scan(&clientID, &passwordHash)
	if err == sql.ErrNoRows {
		writeErrorCode(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный телефон или пароль")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !passwordHash.Valid || bcrypt.CompareHashAndPassword([]byte(passwordHash.String), []byte(request.Password)) != nil {
		writeErrorCode(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный телефон или пароль")
		return
	}

	response, err := a.authResponse(clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) authResponse(clientID int64) (AuthResponse, error) {
	token, err := a.signToken(clientID, "access", 24*time.Hour)
	if err != nil {
		return AuthResponse{}, err
	}
	refreshToken, err := a.signToken(clientID, "refresh", 30*24*time.Hour)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{ClientID: clientID, Token: token, RefreshToken: refreshToken}, nil
}

func (a *App) signToken(clientID int64, tokenType string, ttl time.Duration) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := authClaims{
		ClientID: clientID,
		Type:     tokenType,
		Exp:      time.Now().Add(ttl).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encoder := base64.RawURLEncoding
	unsigned := encoder.EncodeToString(headerJSON) + "." + encoder.EncodeToString(claimsJSON)
	signature := signHMAC([]byte(a.cfg.JWTSecret), unsigned)
	return unsigned + "." + encoder.EncodeToString(signature), nil
}

func (a *App) parseAccessToken(r *http.Request) (int64, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return 0, authError{status: http.StatusUnauthorized, code: "AUTH_REQUIRED", message: "Требуется авторизация"}
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_AUTH_HEADER", message: "Некорректный заголовок авторизации"}
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}

	unsigned := parts[0] + "." + parts[1]
	expected := signHMAC([]byte(a.cfg.JWTSecret), unsigned)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actual, expected) {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}
	var claims authClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}
	if claims.Type != "access" {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN_TYPE", message: "Некорректный тип токена"}
	}
	if claims.Exp < time.Now().Unix() {
		return 0, authError{status: http.StatusUnauthorized, code: "TOKEN_EXPIRED", message: "Срок действия токена истек"}
	}
	if claims.ClientID <= 0 {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}
	return claims.ClientID, nil
}

func (a *App) requireClient(r *http.Request, expectedClientID int64) error {
	clientID, err := a.parseAccessToken(r)
	if err != nil {
		return err
	}
	if clientID != expectedClientID {
		return authError{status: http.StatusForbidden, code: "ACCESS_DENIED", message: "Доступ запрещен"}
	}
	return nil
}

func (a *App) clientIDFromRequest(r *http.Request, requestedClientID int64) (int64, bool, error) {
	if a.isAdminRequest(r) {
		if requestedClientID <= 0 {
			return 0, true, authError{status: http.StatusBadRequest, code: "CLIENT_ID_REQUIRED", message: "clientId обязателен"}
		}
		return requestedClientID, true, nil
	}

	clientID, err := a.parseAccessToken(r)
	if err != nil {
		return 0, false, err
	}
	if requestedClientID > 0 && requestedClientID != clientID {
		return 0, false, authError{status: http.StatusForbidden, code: "CLIENT_ID_MISMATCH", message: "clientId не совпадает с токеном"}
	}
	return clientID, false, nil
}

func (a *App) requireClientOrAdmin(r *http.Request, expectedClientID int64) (bool, error) {
	if a.isAdminRequest(r) {
		return true, nil
	}
	return false, a.requireClient(r, expectedClientID)
}

func (a *App) isAdminRequest(r *http.Request) bool {
	if a.cfg.AdminToken == "" {
		return false
	}

	token := strings.TrimSpace(r.Header.Get("X-API-Token"))
	if token == "" {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(authorization, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		}
	}
	if token == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(a.cfg.AdminToken)) == 1
}

func writeAuthError(w http.ResponseWriter, err error) {
	var authErr authError
	if errors.As(err, &authErr) {
		code := authErr.code
		if code == "" {
			code = "AUTH_ERROR"
		}
		writeErrorCode(w, authErr.status, code, authErr.message)
		return
	}
	writeError(w, http.StatusUnauthorized, err.Error())
}

func writeValidationError(w http.ResponseWriter, err error) {
	if code, ok := validationCode(err); ok {
		writeErrorCode(w, http.StatusBadRequest, code, err.Error())
		return
	}
	writeError(w, http.StatusBadRequest, err.Error())
}

func (a *App) refreshToken(w http.ResponseWriter, r *http.Request) {
	if !a.authLimiter.Allow(rateLimitKey(r, "refresh")) {
		writeErrorCode(w, http.StatusTooManyRequests, "RATE_LIMITED", "Слишком много попыток обновления токена")
		return
	}

	var request struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный JSON")
		return
	}

	clientID, err := a.parseRefreshToken(request.RefreshToken)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	response, err := a.authResponse(clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) parseRefreshToken(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}

	unsigned := parts[0] + "." + parts[1]
	expected := signHMAC([]byte(a.cfg.JWTSecret), unsigned)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actual, expected) {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}
	var claims authClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}
	if claims.Type != "refresh" {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN_TYPE", message: "Некорректный тип токена"}
	}
	if claims.Exp < time.Now().Unix() {
		return 0, authError{status: http.StatusUnauthorized, code: "TOKEN_EXPIRED", message: "Срок действия токена истек"}
	}
	if claims.ClientID <= 0 {
		return 0, authError{status: http.StatusUnauthorized, code: "INVALID_TOKEN", message: "Некорректный токен"}
	}
	return claims.ClientID, nil
}

func signHMAC(secret []byte, value string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return mac.Sum(nil)
}
