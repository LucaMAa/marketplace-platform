package service

import (
	"context"
	"fmt"
	"log"
	"marketplace-platform/models"
	"marketplace-platform/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

// LearningService analyses historical request data to surface insights
// such as "this product was recently requested near you" or merchant suggestion hints.
type LearningService struct {
	requestRepo  *repository.RequestRepository
	redisClient  *redis.Client
	esService    *ElasticsearchService
}

func NewLearningService(
	requestRepo *repository.RequestRepository,
	redisClient *redis.Client,
	esService *ElasticsearchService,
) *LearningService {
	return &LearningService{
		requestRepo: requestRepo,
		redisClient: redisClient,
		esService:   esService,
	}
}

func (s *LearningService) RecordRequest(ctx context.Context, request *models.Request) {
	pipe := s.redisClient.Pipeline()

	popularityKey := fmt.Sprintf("learn:product:popularity:%s", request.ProductCode)
	pipe.Incr(ctx, popularityKey)
	pipe.Expire(ctx, popularityKey, 30*24*time.Hour)

	latBucket := int(request.Latitude)
	lonBucket := int(request.Longitude)
	geoKey := fmt.Sprintf("learn:product:geo:%s:%d:%d", request.ProductCode, latBucket, lonBucket)
	pipe.Incr(ctx, geoKey)
	pipe.Expire(ctx, geoKey, 30*24*time.Hour)

	hour := time.Now().UTC().Hour()
	hourKey := fmt.Sprintf("learn:product:hour:%s:%d", request.ProductCode, hour)
	pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, 90*24*time.Hour)

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("learning: redis pipeline error: %v", err)
	}
}

func (s *LearningService) RecordMerchantResponse(ctx context.Context, merchantID uint, productCode string, hasProduct bool) {
	score := 0
	if hasProduct {
		score = 1
	}
	key := fmt.Sprintf("learn:merchant:%d:products", merchantID)
	s.redisClient.ZIncrBy(ctx, key, float64(score), productCode)
	s.redisClient.Expire(ctx, key, 90*24*time.Hour)
}

func (s *LearningService) GetProductInsights(ctx context.Context, productCode string, lat, lon float64) (*ProductInsights, error) {
	latBucket := int(lat)
	lonBucket := int(lon)

	popularityKey := fmt.Sprintf("learn:product:popularity:%s", productCode)
	geoKey := fmt.Sprintf("learn:product:geo:%s:%d:%d", productCode, latBucket, lonBucket)

	globalCount, _ := s.redisClient.Get(ctx, popularityKey).Int64()
	localCount, _ := s.redisClient.Get(ctx, geoKey).Int64()

	hourProfile := make(map[int]int64, 24)
	for h := 0; h < 24; h++ {
		hourKey := fmt.Sprintf("learn:product:hour:%s:%d", productCode, h)
		v, _ := s.redisClient.Get(ctx, hourKey).Int64()
		hourProfile[h] = v
	}

	return &ProductInsights{
		ProductCode:    productCode,
		GlobalRequests: globalCount,
		LocalRequests:  localCount,
		HourProfile:    hourProfile,
	}, nil
}

func (s *LearningService) SuggestMerchantsForProduct(ctx context.Context, productCode string, candidateIDs []uint) []uint {
	type scored struct {
		id    uint
		score float64
	}

	var results []scored
	for _, id := range candidateIDs {
		key := fmt.Sprintf("learn:merchant:%d:products", id)
		sc, err := s.redisClient.ZScore(ctx, key, productCode).Result()
		if err != nil {
			sc = 0
		}
		results = append(results, scored{id, sc})
	}

	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	out := make([]uint, len(results))
	for i, r := range results {
		out[i] = r.id
	}
	return out
}

func (s *LearningService) GetRecentRequests(limit int) ([]models.Request, error) {
	return s.requestRepo.GetRecentRequests(limit)
}

type ProductInsights struct {
	ProductCode    string         `json:"product_code"`
	GlobalRequests int64          `json:"global_requests"`
	LocalRequests  int64          `json:"local_requests"`
	HourProfile    map[int]int64  `json:"hour_profile"`
}
