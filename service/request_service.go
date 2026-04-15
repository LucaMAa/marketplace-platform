package service

import (
	"context"
	"errors"
	"marketplace-platform/models"
	"marketplace-platform/repository"
	"time"

	"gorm.io/datatypes"
)

type RequestService struct {
	requestRepo    *repository.RequestRepository
	geoService     *GeoService
	redisGeoService *RedisGeoService
	elasticsearchService *ElasticsearchService
	notificationService *NotificationService
}

func NewRequestService(
	requestRepo *repository.RequestRepository,
	geoService *GeoService,
	redisGeoService *RedisGeoService,
	elasticsearchService *ElasticsearchService,
	notificationService *NotificationService,
) *RequestService {
	return &RequestService{
		requestRepo:    requestRepo,
		geoService:     geoService,
		redisGeoService: redisGeoService,
		elasticsearchService: elasticsearchService,
		notificationService: notificationService,
	}
}

// CreateRequest crea una nuova richiesta e la notifica agli esercenti vicini
func (s *RequestService) CreateRequest(ctx context.Context, userID uint, productCode, productName string, quantity int, latitude, longitude float64, radiusKm int) (*models.Request, error) {
	request := &models.Request{
		UserID:      userID,
		ProductCode: productCode,
		ProductName: productName,
		Quantity:    quantity,
		Latitude:    latitude,
		Longitude:   longitude,
		Radius:      radiusKm,
		Status:      models.RequestPending,
		ExpiresAt:   time.Now().Add(2 * time.Hour), // Scade dopo 2 ore
		Metadata:    datatypes.JSONMap{},
	}

	// Salva la richiesta
	if err := s.requestRepo.Create(request); err != nil {
		return nil, err
	}

	// Indexa su Elasticsearch per future ricerche
	go s.elasticsearchService.IndexRequest(ctx, request)

	// Archivia in Redis per ricerche rapide per prodotto
	go s.redisGeoService.StoreRequestKey(ctx, request.ID, productCode)

	// Notifica gli esercenti vicini (in modo asincrono)
	go s.notifyNearbyMerchants(ctx, request)

	return request, nil
}

// notifyNearbyMerchants notifica tutti gli esercenti nelle vicinanze
func (s *RequestService) notifyNearbyMerchants(ctx context.Context, request *models.Request) error {
	// Trova gli esercenti vicini (senza tipo specifico, tutti)
	merchants, err := s.geoService.FindNearestMerchants(
		request.Latitude,
		request.Longitude,
		float64(request.Radius),
		nil, // Tutti i tipi
		100,  // Massimo 100 esercenti per richiesta
	)

	if err != nil {
		return err
	}

	for _, merchant := range merchants {
		// Aggiungi l'esercente alla richiesta
		if err := s.requestRepo.AddMerchant(request.ID, merchant.ID); err != nil {
			continue
		}

		// Crea notifica
		notification := &models.MerchantNotification{
			MerchantID: merchant.ID,
			RequestID:  request.ID,
			Type:       models.NotificationTypeNewRequest,
			IsRead:     false,
		}

		if err := s.notificationService.CreateNotification(ctx, notification); err != nil {
			continue
		}

		// Invia notifica via WebSocket (vedi ws_manager.go)
		s.notificationService.BroadcastToMerchant(merchant.ID, "new_request", map[string]interface{}{
			"request_id":   request.ID,
			"product_name": request.ProductName,
			"quantity":     request.Quantity,
			"distance":     s.geoService.CalculateDistance(request.Latitude, request.Longitude, merchant.Latitude, merchant.Longitude),
		})
	}

	return nil
}

func (s *RequestService) GetRequest(id uint) (*models.Request, error) {
	return s.requestRepo.GetByID(id)
}

func (s *RequestService) GetUserRequests(userID uint) ([]models.Request, error) {
	return s.requestRepo.GetByUserID(userID)
}

func (s *RequestService) UpdateRequestStatus(id uint, status models.RequestStatus) error {
	request, err := s.requestRepo.GetByID(id)
	if err != nil {
		return err
	}
	if request == nil {
		return errors.New("richiesta non trovata")
	}

	request.Status = status
	return s.requestRepo.Update(request)
}

func (s *RequestService) GetLearningData(productCode string) (map[string]interface{}, error) {	
	data := map[string]interface{}{
		"product_code":     productCode,
		"recent_requests":  5,
		"response_rate":    0.75,
		"avg_response_time": "5 minutes",
	}

	return data, nil
}
