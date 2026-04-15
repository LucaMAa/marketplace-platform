package handlers

import (
	"encoding/json"
	"marketplace-platform/middleware"
	"marketplace-platform/models"
	"marketplace-platform/repository"
	"marketplace-platform/service"
	"net/http"
	"time"
)

type ChatHandler struct {
	chatRepo        *repository.ChatRepository
	requestService  *service.RequestService
	notifService    *service.NotificationService
	learningService *service.LearningService
}

func NewChatHandler(
	chatRepo *repository.ChatRepository,
	requestService *service.RequestService,
	notifService *service.NotificationService,
	learningService *service.LearningService,
) *ChatHandler {
	return &ChatHandler{
		chatRepo:        chatRepo,
		requestService:  requestService,
		notifService:    notifService,
		learningService: learningService,
	}
}

// GET /requests/:id/messages
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	requestID, err := pathUintParam(r, "id")
	if err != nil {
		jsonError(w, "invalid request id", http.StatusBadRequest)
		return
	}

	role := middleware.GetRole(r)
	userID := middleware.GetUserID(r)

	var messages []models.ChatMessage

	if role == "merchant" {
		messages, err = h.chatRepo.GetByRequestIDAndMerchant(requestID, userID)
	} else {
		messages, err = h.chatRepo.GetByRequestID(requestID)
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, messages)
}

// POST /requests/:id/messages  — merchant replies to a request
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	requestID, err := pathUintParam(r, "id")
	if err != nil {
		jsonError(w, "invalid request id", http.StatusBadRequest)
		return
	}

	role := middleware.GetRole(r)
	senderID := middleware.GetUserID(r)

	var req struct {
		Content    string   `json:"content"`
		HasProduct bool     `json:"has_product"`
		CanOrder   bool     `json:"can_order"`
		Price      *float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		jsonError(w, "content is required", http.StatusBadRequest)
		return
	}

	// Fetch the parent request to get the user_id for notification
	request, err := h.requestService.GetRequest(requestID)
	if err != nil || request == nil {
		jsonError(w, "request not found", http.StatusNotFound)
		return
	}

	msg := &models.ChatMessage{
		RequestID:  requestID,
		Content:    req.Content,
		HasProduct: req.HasProduct,
		CanOrder:   req.CanOrder,
		Price:      req.Price,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if role == "merchant" {
		msg.MerchantID = &senderID
		msg.Type = models.MessageTypeResponse

		// Record learning signal
		go h.learningService.RecordMerchantResponse(r.Context(), senderID, request.ProductCode, req.HasProduct)

		// Notify the user in real-time
		h.notifService.BroadcastToUser(request.UserID, "new_message", map[string]interface{}{
			"request_id":  requestID,
			"merchant_id": senderID,
			"content":     req.Content,
			"has_product": req.HasProduct,
			"can_order":   req.CanOrder,
			"price":       req.Price,
		})

		// Also notify merchant of their own success (optional echo)
		notif := &models.MerchantNotification{
			MerchantID: senderID,
			RequestID:  requestID,
			Type:       models.NotificationTypeMessage,
			IsRead:     false,
		}
		_ = h.notifService.CreateNotification(r.Context(), notif)

	} else {
		// User side: sending follow-up to a specific merchant
		msg.UserID = &senderID
		msg.Type = models.MessageTypeChat

		merchantID, err := pathUintParam(r, "merchant_id")
		if err == nil && merchantID != 0 {
			msg.MerchantID = &merchantID
			h.notifService.BroadcastToMerchant(merchantID, "new_message", map[string]interface{}{
				"request_id": requestID,
				"content":    req.Content,
			})
		}
	}

	if err := h.chatRepo.Create(msg); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, msg)
}
