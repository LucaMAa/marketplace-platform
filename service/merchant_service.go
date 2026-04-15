package service

import (
	"context"
	"errors"
	"marketplace-platform/models"
	"marketplace-platform/repository"
)

type MerchantService struct {
	merchantRepo  *repository.MerchantRepository
	inventoryRepo *repository.InventoryRepository
	productRepo   *repository.ProductRepository
	redisGeo      *RedisGeoService
	esService     *ElasticsearchService
}

func NewMerchantService(
	merchantRepo *repository.MerchantRepository,
	inventoryRepo *repository.InventoryRepository,
	productRepo *repository.ProductRepository,
	redisGeo *RedisGeoService,
	esService *ElasticsearchService,
) *MerchantService {
	return &MerchantService{
		merchantRepo:  merchantRepo,
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		redisGeo:      redisGeo,
		esService:     esService,
	}
}

func (s *MerchantService) GetByID(id uint) (*models.Merchant, error) {
	return s.merchantRepo.GetByID(id)
}

func (s *MerchantService) UpdateProfile(merchant *models.Merchant) error {
	return s.merchantRepo.Update(merchant)
}

func (s *MerchantService) SetOnline(ctx context.Context, merchant *models.Merchant) error {
	if err := s.merchantRepo.UpdateLastActive(merchant.ID); err != nil {
		return err
	}
	return s.redisGeo.IndexMerchantLocation(ctx, merchant)
}

func (s *MerchantService) SetOffline(ctx context.Context, merchant *models.Merchant) error {
	return s.redisGeo.RemoveMerchantLocation(ctx, merchant)
}

func (s *MerchantService) GetInventory(merchantID uint) ([]models.Inventory, error) {
	return s.inventoryRepo.GetByMerchantID(merchantID)
}

func (s *MerchantService) UpsertInventory(ctx context.Context, merchantID, productID uint, quantity int, canOrder bool, price *float64) (*models.Inventory, error) {
	inv, err := s.inventoryRepo.GetByMerchantAndProduct(merchantID, productID)
	if err != nil {
		return nil, err
	}

	if inv == nil {
		product, err := s.productRepo.GetByID(productID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, errors.New("product not found")
		}

		inv = &models.Inventory{
			MerchantID: merchantID,
			ProductID:  productID,
		}
	}

	inv.Quantity = quantity
	inv.CanOrder = canOrder
	inv.Price = price

	if inv.ID == 0 {
		if err := s.inventoryRepo.Create(inv); err != nil {
			return nil, err
		}
	} else {
		if err := s.inventoryRepo.Update(inv); err != nil {
			return nil, err
		}
	}

	return inv, nil
}

func (s *MerchantService) GetNearby(lat, lon, radiusKm float64, merchantType *models.MerchantType) ([]models.Merchant, error) {
	return s.merchantRepo.GetNearby(lat, lon, radiusKm, merchantType)
}
