package handlers

import (
	"net/http"

	"marketplace-platform/middleware"
	"marketplace-platform/service"
)

type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler(ns *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: ns}
}

// GET /notifications  — merchant only
func (h *NotificationHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	items, err := h.notifService.GetByMerchant(merchantID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}

// GET /notifications/unread
func (h *NotificationHandler) GetUnread(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	items, err := h.notifService.GetUnread(merchantID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}

// PATCH /notifications/:id/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := pathUintParam(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.notifService.MarkAsRead(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

// PATCH /notifications/read-all
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	merchantID := middleware.GetUserID(r)
	if err := h.notifService.MarkAllAsRead(merchantID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}
