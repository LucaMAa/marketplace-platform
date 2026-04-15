package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"marketplace-platform/middleware"
	"marketplace-platform/models"
	"marketplace-platform/service"
)

type RequestHandler struct {
	requestService  *service.RequestService
	learningService *service.LearningService
}

func NewRequestHandler(rs *service.RequestService, ls *service.LearningService) *RequestHandler {
	return &RequestHandler{requestService: rs, learningService: ls}
}

// POST /requests
func (h *RequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		ProductCode string  `json:"product_code"`
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		RadiusKm    int     `json:"radius_km"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.RadiusKm <= 0 {
		req.RadiusKm = 5
	}
	request, err := h.requestService.CreateRequest(
		r.Context(), userID,
		req.ProductCode, req.ProductName, req.Quantity,
		req.Latitude, req.Longitude, req.RadiusKm,
	)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go h.learningService.RecordRequest(r.Context(), request)
	jsonOK(w, request)
}

func (h *RequestHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	id, err := pathUintParam(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	request, err := h.requestService.GetRequest(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if request == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, request)
}

func (h *RequestHandler) GetMyRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	requests, err := h.requestService.GetUserRequests(userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, requests)
}

func (h *RequestHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathUintParam(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req struct {
		Status models.RequestStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.requestService.UpdateRequestStatus(id, req.Status); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "updated"})
}

func (h *RequestHandler) GetInsights(w http.ResponseWriter, r *http.Request) {
	productCode := r.URL.Query().Get("product_code")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	insights, err := h.learningService.GetProductInsights(r.Context(), productCode, lat, lon)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, insights)
}

func pathUintParam(r *http.Request, name string) (uint, error) {
	raw := r.PathValue(name)
	if raw == "" {
		raw = r.URL.Query().Get(name)
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	return uint(v), err
}
