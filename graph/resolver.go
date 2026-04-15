package graph

import (
	"context"
	"encoding/json"
	"errors"
	"marketplace-platform/middleware"
	"marketplace-platform/models"
	"marketplace-platform/repository"
	"marketplace-platform/service"
	"net/http"
	"strconv"
)




type Resolver struct {
	AuthService         *service.AuthService
	RequestService      *service.RequestService
	MerchantService     *service.MerchantService
	NotifService        *service.NotificationService
	LearningService     *service.LearningService
	ESService           *service.ElasticsearchService
	ChatRepo            *repository.ChatRepository
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   interface{}      `json:"data,omitempty"`
	Errors []GraphQLError   `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}


func (r *Resolver) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var gqlReq GraphQLRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			writeGQLError(w, "invalid request body")
			return
		}

		ctx := req.Context()
		data, err := r.dispatch(ctx, gqlReq)
		if err != nil {
			json.NewEncoder(w).Encode(GraphQLResponse{
				Errors: []GraphQLError{{Message: err.Error()}},
			})
			return
		}
		json.NewEncoder(w).Encode(GraphQLResponse{Data: data})
	}
}

func (r *Resolver) dispatch(ctx context.Context, req GraphQLRequest) (interface{}, error) {
	vars := req.Variables

	switch operationName(req) {
	// ── Auth ──────────────────────────────────────────────
	case "registerUser":
		user, token, err := r.AuthService.RegisterUser(
			str(vars, "email"), str(vars, "password"), str(vars, "name"),
			f64(vars, "latitude"), f64(vars, "longitude"), str(vars, "address"),
		)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"token": token, "user": user}, nil

	case "loginUser":
		user, token, err := r.AuthService.LoginUser(str(vars, "email"), str(vars, "password"))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"token": token, "user": user}, nil

	case "registerMerchant":
		mt := models.MerchantType(str(vars, "type"))
		merchant, token, err := r.AuthService.RegisterMerchant(
			str(vars, "email"), str(vars, "password"), str(vars, "businessName"), mt,
			f64(vars, "latitude"), f64(vars, "longitude"),
			str(vars, "address"), str(vars, "city"), str(vars, "phone"),
		)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"token": token, "merchant": merchant}, nil

	case "loginMerchant":
		merchant, token, err := r.AuthService.LoginMerchant(str(vars, "email"), str(vars, "password"))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"token": token, "merchant": merchant}, nil

	// ── Requests ──────────────────────────────────────────
	case "createRequest":
		userID := middleware.GetUserIDFromCtx(ctx)
		if userID == 0 {
			return nil, errors.New("unauthorized")
		}
		radius := int(f64(vars, "radiusKm"))
		if radius == 0 {
			radius = 5
		}
		request, err := r.RequestService.CreateRequest(
			ctx, userID,
			str(vars, "productCode"), str(vars, "productName"), int(f64(vars, "quantity")),
			f64(vars, "latitude"), f64(vars, "longitude"), radius,
		)
		if err != nil {
			return nil, err
		}
		go r.LearningService.RecordRequest(ctx, request)
		return request, nil

	case "myRequests":
		userID := middleware.GetUserIDFromCtx(ctx)
		return r.RequestService.GetUserRequests(userID)

	case "request":
		id := toUint(vars, "id")
		return r.RequestService.GetRequest(id)

	// ── Chat ──────────────────────────────────────────────
	case "requestMessages":
		requestID := toUint(vars, "requestID")
		role := middleware.GetRoleFromCtx(ctx)
		senderID := middleware.GetUserIDFromCtx(ctx)
		if role == "merchant" {
			return r.ChatRepo.GetByRequestIDAndMerchant(requestID, senderID)
		}
		return r.ChatRepo.GetByRequestID(requestID)

	case "sendMessage":
		return r.handleSendMessage(ctx, vars)

	// ── Notifications ─────────────────────────────────────
	case "myNotifications":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		return r.NotifService.GetByMerchant(merchantID)

	case "unreadNotifications":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		return r.NotifService.GetUnread(merchantID)

	case "markNotificationRead":
		id := toUint(vars, "id")
		err := r.NotifService.MarkAsRead(id)
		return err == nil, err

	case "markAllNotificationsRead":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		err := r.NotifService.MarkAllAsRead(merchantID)
		return err == nil, err

	// ── Merchant ──────────────────────────────────────────
	case "merchantMe":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		return r.MerchantService.GetByID(merchantID)

	case "goOnline":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		merchant, err := r.MerchantService.GetByID(merchantID)
		if err != nil {
			return false, err
		}
		err = r.MerchantService.SetOnline(ctx, merchant)
		return err == nil, err

	case "goOffline":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		merchant, err := r.MerchantService.GetByID(merchantID)
		if err != nil {
			return false, err
		}
		err = r.MerchantService.SetOffline(ctx, merchant)
		return err == nil, err

	case "myInventory":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		return r.MerchantService.GetInventory(merchantID)

	case "upsertInventory":
		merchantID := middleware.GetUserIDFromCtx(ctx)
		productID := toUint(vars, "productID")
		quantity := int(f64(vars, "quantity"))
		canOrder := boolVar(vars, "canOrder")
		var price *float64
		if p, ok := vars["price"].(float64); ok {
			price = &p
		}
		return r.MerchantService.UpsertInventory(ctx, merchantID, productID, quantity, canOrder, price)

	// ── Search ────────────────────────────────────────────
	case "searchProducts":
		return r.ESService.SearchProducts(ctx, str(vars, "query"))

	case "productInsights":
		return r.LearningService.GetProductInsights(ctx, str(vars, "productCode"), f64(vars, "lat"), f64(vars, "lon"))

	default:
		return nil, errors.New("unknown operation: " + operationName(req))
	}
}

func (r *Resolver) handleSendMessage(ctx context.Context, vars map[string]interface{}) (interface{}, error) {
	role := middleware.GetRoleFromCtx(ctx)
	senderID := middleware.GetUserIDFromCtx(ctx)
	requestID := toUint(vars, "requestID")

	request, err := r.RequestService.GetRequest(requestID)
	if err != nil || request == nil {
		return nil, errors.New("request not found")
	}

	hasProduct := boolVar(vars, "hasProduct")
	canOrder := boolVar(vars, "canOrder")
	var price *float64
	if p, ok := vars["price"].(float64); ok {
		price = &p
	}

	msg := &models.ChatMessage{
		RequestID:  requestID,
		Content:    str(vars, "content"),
		HasProduct: hasProduct,
		CanOrder:   canOrder,
		Price:      price,
	}

	if role == "merchant" {
		msg.MerchantID = &senderID
		msg.Type = models.MessageTypeResponse
		go r.LearningService.RecordMerchantResponse(ctx, senderID, request.ProductCode, hasProduct)
		r.NotifService.BroadcastToUser(request.UserID, "new_message", map[string]interface{}{
			"request_id":  requestID,
			"merchant_id": senderID,
			"content":     msg.Content,
			"has_product": hasProduct,
		})
	} else {
		msg.UserID = &senderID
		msg.Type = models.MessageTypeChat
	}

	if err := r.ChatRepo.Create(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func operationName(req GraphQLRequest) string {
	if req.OperationName != "" {
		return req.OperationName
	}
	// Very naive extraction — replace with proper parser in production
	return req.OperationName
}

func str(vars map[string]interface{}, key string) string {
	v, _ := vars[key].(string)
	return v
}

func f64(vars map[string]interface{}, key string) float64 {
	switch v := vars[key].(type) {
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

func toUint(vars map[string]interface{}, key string) uint {
	switch v := vars[key].(type) {
	case float64:
		return uint(v)
	case string:
		u, _ := strconv.ParseUint(v, 10, 64)
		return uint(u)
	}
	return 0
}

func boolVar(vars map[string]interface{}, key string) bool {
	v, _ := vars[key].(bool)
	return v
}

func writeGQLError(w http.ResponseWriter, msg string) {
	json.NewEncoder(w).Encode(GraphQLResponse{
		Errors: []GraphQLError{{Message: msg}},
	})
}
