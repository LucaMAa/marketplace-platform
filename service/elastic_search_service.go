package service

import (
	"bytes"
	"context"
	"encoding/json"
	"marketplace-platform/models"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchService struct {
	client *elasticsearch.Client
}

func NewElasticsearchService(client *elasticsearch.Client) *ElasticsearchService {
	return &ElasticsearchService{client: client}
}

// IndexProduct indexa un prodotto in Elasticsearch
func (s *ElasticsearchService) IndexProduct(ctx context.Context, product *models.Product) error {
	doc := map[string]interface{}{
		"id":          product.ID,
		"name":        product.Name,
		"description": product.Description,
		"code":        product.Code,
		"category":    product.Category,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := &bytes.Buffer{}
	req.Write(data)

	res, err := s.client.Index(
		"products",
		req,
		s.client.Index.WithContext(ctx),
		s.client.Index.WithDocumentID(product.Code),
	)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

// SearchProducts ricerca prodotti con Elasticsearch
func (s *ElasticsearchService) SearchProducts(ctx context.Context, query string) ([]models.Product, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name^3", "code^2", "description", "category"},
			},
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex("products"),
		s.client.Search.WithBody(bytes.NewReader(body)),
	)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				Source models.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	var products []models.Product
	for _, hit := range result.Hits.Hits {
		products = append(products, hit.Source)
	}

	return products, nil
}

// IndexRequest indexa una richiesta per ricerche future (learning)
func (s *ElasticsearchService) IndexRequest(ctx context.Context, request *models.Request) error {
	doc := map[string]interface{}{
		"id":           request.ID,
		"user_id":      request.UserID,
		"product_code": request.ProductCode,
		"product_name": request.ProductName,
		"quantity":     request.Quantity,
		"latitude":     request.Latitude,
		"longitude":    request.Longitude,
		"radius":       request.Radius,
		"created_at":   request.CreatedAt,
		"status":       request.Status,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := s.client.Index(
		"requests",
		bytes.NewReader(data),
		s.client.Index.WithContext(ctx),
		s.client.Index.WithDocumentID(request.ProductCode),
	)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
