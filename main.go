package main

import (
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"

	"marketplace-platform/config"
	"marketplace-platform/graph"
	"marketplace-platform/handlers"
	"marketplace-platform/middleware"
	"marketplace-platform/repository"
	"marketplace-platform/service"
	"marketplace-platform/utils"
	ws "marketplace-platform/websocket"
)

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg := config.Load()
	utils.SetJWTSecret(cfg.JWTSecret)

	// ── Infrastructure ────────────────────────────────────────────────────────
	db := config.InitDB(cfg)
	redisClient := config.InitRedis(cfg)

	esCfg := elasticsearch.Config{Addresses: []string{cfg.ESUrl}}
	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Fatalf("elasticsearch client: %v", err)
	}

	// ── WebSocket manager ─────────────────────────────────────────────────────
	wsManager := ws.NewManager()

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	merchantRepo := repository.NewMerchantRepository(db)
	productRepo := repository.NewProductRepository(db)
	requestRepo := repository.NewRequestRepository(db)
	chatRepo := repository.NewChatRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	// ── Services ──────────────────────────────────────────────────────────────
	authService := service.NewAuthService(userRepo, merchantRepo)
	redisGeoService := service.NewRedisGeoService(redisClient, merchantRepo)
	esService := service.NewElasticsearchService(esClient)
	geoService := service.NewGeoService(merchantRepo)
	notifService := service.NewNotificationService(notifRepo, wsManager)
	requestService := service.NewRequestService(requestRepo, geoService, redisGeoService, esService, notifService)
	merchantService := service.NewMerchantService(merchantRepo, inventoryRepo, productRepo, redisGeoService, esService)
	learningService := service.NewLearningService(requestRepo, redisClient, esService)

	// ── HTTP Handlers ─────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService)
	requestHandler := handlers.NewRequestHandler(requestService, learningService)
	merchantHandler := handlers.NewMerchantHandler(merchantService)
	chatHandler := handlers.NewChatHandler(chatRepo, requestService, notifService, learningService)
	notifHandler := handlers.NewNotificationHandler(notifService)

	// ── GraphQL resolver ──────────────────────────────────────────────────────
	resolver := &graph.Resolver{
		AuthService:     authService,
		RequestService:  requestService,
		MerchantService: merchantService,
		NotifService:    notifService,
		LearningService: learningService,
		ESService:       esService,
		ChatRepo:        chatRepo,
	}

	// ── Router (stdlib ServeMux, Go 1.22+ path params) ────────────────────────
	mux := http.NewServeMux()

	// Auth (public)
	mux.HandleFunc("POST /auth/register/user", authHandler.RegisterUser)
	mux.HandleFunc("POST /auth/login/user", authHandler.LoginUser)
	mux.HandleFunc("POST /auth/register/merchant", authHandler.RegisterMerchant)
	mux.HandleFunc("POST /auth/login/merchant", authHandler.LoginMerchant)

	// Requests (user)
	mux.HandleFunc("POST /requests", middleware.RequireAuth(requestHandler.CreateRequest))
	mux.HandleFunc("GET /requests/me", middleware.RequireAuth(requestHandler.GetMyRequests))
	mux.HandleFunc("GET /requests/{id}", middleware.RequireAuth(requestHandler.GetRequest))
	mux.HandleFunc("PATCH /requests/{id}/status", middleware.RequireAuth(requestHandler.UpdateStatus))
	mux.HandleFunc("GET /insights", middleware.RequireAuth(requestHandler.GetInsights))

	// Chat
	mux.HandleFunc("GET /requests/{id}/messages", middleware.RequireAuth(chatHandler.GetMessages))
	mux.HandleFunc("POST /requests/{id}/messages", middleware.RequireAuth(chatHandler.SendMessage))

	// Merchant
	mux.HandleFunc("GET /merchants/me", middleware.RequireRole("merchant", merchantHandler.GetProfile))
	mux.HandleFunc("PATCH /merchants/me", middleware.RequireRole("merchant", merchantHandler.UpdateProfile))
	mux.HandleFunc("POST /merchants/me/online", middleware.RequireRole("merchant", merchantHandler.GoOnline))
	mux.HandleFunc("POST /merchants/me/offline", middleware.RequireRole("merchant", merchantHandler.GoOffline))
	mux.HandleFunc("GET /merchants/me/inventory", middleware.RequireRole("merchant", merchantHandler.GetInventory))
	mux.HandleFunc("POST /merchants/me/inventory", middleware.RequireRole("merchant", merchantHandler.UpsertInventory))

	// Notifications (merchant)
	mux.HandleFunc("GET /notifications", middleware.RequireRole("merchant", notifHandler.GetAll))
	mux.HandleFunc("GET /notifications/unread", middleware.RequireRole("merchant", notifHandler.GetUnread))
	mux.HandleFunc("PATCH /notifications/{id}/read", middleware.RequireRole("merchant", notifHandler.MarkRead))
	mux.HandleFunc("PATCH /notifications/read-all", middleware.RequireRole("merchant", notifHandler.MarkAllRead))

	// WebSocket endpoints
	mux.HandleFunc("GET /ws/merchant", func(w http.ResponseWriter, r *http.Request) {
		claims, err := extractClaims(r)
		if err != nil || claims.Role != "merchant" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		wsManager.RegisterMerchant(w, r, claims.ID)
	})
	mux.HandleFunc("GET /ws/user", func(w http.ResponseWriter, r *http.Request) {
		claims, err := extractClaims(r)
		if err != nil || claims.Role != "user" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		wsManager.RegisterUser(w, r, claims.ID)
	})

	// GraphQL
	mux.Handle("POST /graphql", middleware.AuthMiddleware(resolver.Handler()))

	// Global auth middleware
	handler := middleware.AuthMiddleware(mux)

	log.Printf("🚀 Server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func extractClaims(r *http.Request) (*utils.Claims, error) {
	token := r.URL.Query().Get("token")
	return utils.ValidateToken(token)
}
