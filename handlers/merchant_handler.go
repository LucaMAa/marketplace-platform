package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"marketplace-platform/middleware"
	"marketplace-platform/service"
)

type MerchantHandler struct {
	merchantService *service.MerchantService
	wsCtx           context.Context
}

func NewMerchantHandler(ms *service.MerchantService) *MerchantHandler {
	return &MerchantHandler{merchantService: ms, wsCtx: context.Background()}
}

// GET /merchants/me
func (h *MerchantHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	merchant, err := h.merchantService.GetByID(merchantID)
	if err != nil || merchant == nil {
		jsonError(w, "merchant not found", http.StatusNotFound)
		return
	}
	jsonOK(w, merchant)
}

// PATCH /merchants/me
func (h *MerchantHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	merchant, err := h.merchantService.GetByID(merchantID)
	if err != nil || merchant == nil {
		jsonError(w, "merchant not found", http.StatusNotFound)
		return
	}

	var req struct {
		BusinessName string   `json:"business_name"`
		Address      string   `json:"address"`
		City         string   `json:"city"`
		Phone        string   `json:"phone"`
		Website      *string  `json:"website"`
		Latitude     *float64 `json:"latitude"`
		Longitude    *float64 `json:"longitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.BusinessName != "" {
		merchant.BusinessName = req.BusinessName
	}
	if req.Address != "" {
		merchant.Address = req.Address
	}
	if req.City != "" {
		merchant.City = req.City
	}
	if req.Phone != "" {
		merchant.Phone = req.Phone
	}
	if req.Website != nil {
		merchant.Website = req.Website
	}
	if req.Latitude != nil {
		merchant.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		merchant.Longitude = *req.Longitude
	}

	if err := h.merchantService.UpdateProfile(merchant); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, merchant)
}

// POST /merchants/me/online  — merchant marks themselves online, gets indexed in Redis geo
func (h *MerchantHandler) GoOnline(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	merchant, err := h.merchantService.GetByID(merchantID)
	if err != nil || merchant == nil {
		jsonError(w, "merchant not found", http.StatusNotFound)
		return
	}
	if err := h.merchantService.SetOnline(r.Context(), merchant); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "online"})
}

// POST /merchants/me/offline
func (h *MerchantHandler) GoOffline(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	merchant, err := h.merchantService.GetByID(merchantID)
	if err != nil || merchant == nil {
		jsonError(w, "merchant not found", http.StatusNotFound)
		return
	}
	if err := h.merchantService.SetOffline(r.Context(), merchant); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "offline"})
}

// GET /merchants/me/inventory
func (h *MerchantHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	items, err := h.merchantService.GetInventory(merchantID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}

// POST /merchants/me/inventory
func (h *MerchantHandler) UpsertInventory(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)

	var req struct {
		ProductID uint     `json:"product_id"`
		Quantity  int      `json:"quantity"`
		CanOrder  bool     `json:"can_order"`
		Price     *float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}

	inv, err := h.merchantService.UpsertInventory(r.Context(), merchantID, req.ProductID, req.Quantity, req.CanOrder, req.Price)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, inv)
}
