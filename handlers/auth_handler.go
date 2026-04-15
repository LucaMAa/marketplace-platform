package handlers

import (
	"encoding/json"
	"net/http"

	"marketplace-platform/models"
	"marketplace-platform/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		Name      string  `json:"name"`
		Latitude  float64  `json:"latitude"`
		Longitude float64  `json:"longitude"`
		Address   string  `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		jsonError(w, "email, password and name are required", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.RegisterUser(req.Email, req.Password, req.Name, req.Latitude, req.Longitude, req.Address)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]interface{}{"user": user, "token": token})
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.LoginUser(req.Email, req.Password)
	if err != nil {
		jsonError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	jsonOK(w, map[string]interface{}{"user": user, "token": token})
}

func (h *AuthHandler) RegisterMerchant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email        string              `json:"email"`
		Password     string              `json:"password"`
		BusinessName string              `json:"business_name"`
		Type         models.MerchantType `json:"type"`
		Latitude     float64             `json:"latitude"`
		Longitude    float64             `json:"longitude"`
		Address      string              `json:"address"`
		City         string              `json:"city"`
		Phone        string              `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}

	merchant, token, err := h.authService.RegisterMerchant(
		req.Email, req.Password, req.BusinessName, req.Type,
		req.Latitude, req.Longitude, req.Address, req.City, req.Phone,
	)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]interface{}{"merchant": merchant, "token": token})
}

// POST /auth/login/merchant
func (h *AuthHandler) LoginMerchant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}

	merchant, token, err := h.authService.LoginMerchant(req.Email, req.Password)
	if err != nil {
		jsonError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	jsonOK(w, map[string]interface{}{"merchant": merchant, "token": token})
}

// helpers shared across handlers package
func jsonOK(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
