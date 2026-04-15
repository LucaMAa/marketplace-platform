package service

import (
	"context"
	"marketplace-platform/models"
	"marketplace-platform/repository"
	"marketplace-platform/websocket"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
	wsManager *websocket.Manager
}

func NewNotificationService(notifRepo *repository.NotificationRepository, wsManager *websocket.Manager) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
		wsManager: wsManager,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, n *models.MerchantNotification) error {
	return s.notifRepo.Create(n)
}

func (s *NotificationService) GetByMerchant(merchantID uint) ([]models.MerchantNotification, error) {
	return s.notifRepo.GetByMerchantID(merchantID)
}

func (s *NotificationService) GetUnread(merchantID uint) ([]models.MerchantNotification, error) {
	return s.notifRepo.GetUnread(merchantID)
}

func (s *NotificationService) MarkAsRead(id uint) error {
	return s.notifRepo.MarkAsRead(id)
}

func (s *NotificationService) MarkAllAsRead(merchantID uint) error {
	return s.notifRepo.MarkAllAsRead(merchantID)
}

func (s *NotificationService) BroadcastToMerchant(merchantID uint, msgType string, payload interface{}) {
	s.wsManager.BroadcastToMerchant(merchantID, msgType, payload)
}

func (s *NotificationService) BroadcastToUser(userID uint, msgType string, payload interface{}) {
	s.wsManager.BroadcastToUser(userID, msgType, payload)
}
