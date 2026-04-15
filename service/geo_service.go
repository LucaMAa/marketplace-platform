package service

import (
	"marketplace-platform/models"
	"marketplace-platform/repository"
	"math"
)

type GeoService struct {
	merchantRepo *repository.MerchantRepository
}

func NewGeoService(merchantRepo *repository.MerchantRepository) *GeoService {
	return &GeoService{merchantRepo: merchantRepo}
}

func (s *GeoService) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371

	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (s *GeoService) FindNearestMerchants(latitude, longitude, radiusKm float64, merchantType *models.MerchantType, limit int) ([]models.Merchant, error) {
	merchants, err := s.merchantRepo.GetNearby(latitude, longitude, radiusKm, merchantType)
	if err != nil {
		return nil, err
	}

	type merchantWithDist struct {
		merchant models.Merchant
		distance float64
	}

	var withDist []merchantWithDist
	for _, m := range merchants {
		dist := s.CalculateDistance(latitude, longitude, m.Latitude, m.Longitude)
		withDist = append(withDist, merchantWithDist{m, dist})
	}

	// sorta con libreria
	for i := 0; i < len(withDist); i++ {
		for j := i + 1; j < len(withDist); j++ {
			if withDist[j].distance < withDist[i].distance {
				withDist[i], withDist[j] = withDist[j], withDist[i]
			}
		}
	}

	if limit > 0 && len(withDist) > limit {
		withDist = withDist[:limit]
	}

	var result []models.Merchant
	for _, wd := range withDist {
		result = append(result, wd.merchant)
	}

	return result, nil
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
