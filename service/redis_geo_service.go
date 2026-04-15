package service

import (
	"context"
	"fmt"
	"marketplace-platform/models"
	"marketplace-platform/repository"

	"github.com/redis/go-redis/v9"
)

type RedisGeoService struct {
	client            *redis.Client
	merchantRepo      *repository.MerchantRepository
}

func NewRedisGeoService(client *redis.Client, merchantRepo *repository.MerchantRepository) *RedisGeoService {
	return &RedisGeoService{
		client:       client,
		merchantRepo: merchantRepo,
	}
}

// IndexMerchantLocation indexa la posizione dell'esercente in Redis Geo
func (s *RedisGeoService) IndexMerchantLocation(ctx context.Context, merchant *models.Merchant) error {
	key := fmt.Sprintf("merchants:geo:%s", merchant.Type) // merchants:geo:pharmacy, merchants:geo:hardware, ecc
	
	return s.client.GeoAdd(ctx, key, &redis.GeoLocation{
		Name:      fmt.Sprintf("merchant:%d", merchant.ID),
		Longitude: merchant.Longitude,
		Latitude:  merchant.Latitude,
	}).Err()
}

// FindNearbyMerchants trova esercenti nelle vicinanze usando Redis Geo
func (s *RedisGeoService) FindNearbyMerchants(ctx context.Context, latitude, longitude, radiusKm float64, merchantType models.MerchantType) ([]uint, error) {
	key := fmt.Sprintf("merchants:geo:%s", merchantType)
	
	// Redis restituisce i risultati ordinati per distanza
	results := s.client.GeoRadius(ctx, key, longitude, latitude, &redis.GeoRadiusQuery{
		Radius:      radiusKm,
		Unit:        "km",
		WithDist:    true,
		WithCoord:   true,
		Count:       100,
		Sort:        "ASC",
	}).Val()


	var merchantIDs []uint
	for _, result := range results {
		var id uint
		_, err := fmt.Sscanf(result.Name, "merchant:%d", &id)
		if err == nil {
			merchantIDs = append(merchantIDs, id)
		}
	}

	return merchantIDs, nil
}

// RemoveMerchantLocation rimuove la posizione dell'esercente da Redis (quando si disattiva)
func (s *RedisGeoService) RemoveMerchantLocation(ctx context.Context, merchant *models.Merchant) error {
	key := fmt.Sprintf("merchants:geo:%s", merchant.Type)
	return s.client.ZRem(ctx, key, fmt.Sprintf("merchant:%d", merchant.ID)).Err()
}

// CacheMerchantInfo cachizza le info dell'esercente
func (s *RedisGeoService) CacheMerchantInfo(ctx context.Context, merchant *models.Merchant, ttlSeconds int) error {
	key := fmt.Sprintf("merchant:info:%d", merchant.ID)
	data := fmt.Sprintf("%s|%s|%s|%d", merchant.BusinessName, merchant.Address, merchant.Type, merchant.Status)
	return s.client.Set(ctx, key, data, 0).Err() // TTL opzionale
}

// GetCachedMerchantInfo ottiene le info cachizzate
func (s *RedisGeoService) GetCachedMerchantInfo(ctx context.Context, merchantID uint) (string, error) {
	key := fmt.Sprintf("merchant:info:%d", merchantID)
	return s.client.Get(ctx, key).Result()
}

// IndexRequestKeywords indexa le parole chiave della richiesta per Elasticsearch
func (s *RedisGeoService) StoreRequestKey(ctx context.Context, requestID uint, productCode string) error {
	key := fmt.Sprintf("request:product:%s", productCode)
	return s.client.SAdd(ctx, key, requestID).Err()
}

// GetRequestsByProduct ottiene richieste per prodotto
func (s *RedisGeoService) GetRequestsByProduct(ctx context.Context, productCode string) ([]string, error) {
	key := fmt.Sprintf("request:product:%s", productCode)
	return s.client.SMembers(ctx, key).Result()
}
