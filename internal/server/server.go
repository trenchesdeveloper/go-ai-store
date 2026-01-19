package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/events"
	"github.com/trenchesdeveloper/go-ai-store/internal/interfaces"
	"github.com/trenchesdeveloper/go-ai-store/internal/providers"
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
)

type Server struct {
	cfg            *config.Config
	logger         *zerolog.Logger
	store          db.Store
	authService    *services.AuthService
	userService    *services.UserService
	productService *services.ProductService
	uploadService  *services.UploadService
	cartService    *services.CartService
	orderService   *services.OrderService
}

func NewServer(cfg *config.Config, logger *zerolog.Logger, store db.Store) (*Server, error) {
	// Initialize upload provider based on config
	var uploadProvider interfaces.Upload
	switch cfg.Upload.Provider {
	case "s3":
		s3Provider, err := providers.NewS3UploadProvider(providers.S3Config{
			Endpoint:        cfg.AWS.S3Endpoint,
			Region:          cfg.AWS.Region,
			AccessKeyID:     cfg.AWS.AccessKeyID,
			SecretAccessKey: cfg.AWS.SecretAccessKey,
			Bucket:          cfg.AWS.S3Bucket,
		})
		if err != nil {
			return nil, err
		}
		uploadProvider = s3Provider
	default:
		uploadProvider = providers.NewLocalUploadProvider(cfg.Upload.UploadPath)
	}

	// Initialize event publisher
	pub, err := events.NewEventPublisher(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	cartService := services.NewCartService(store)
	return &Server{
		cfg:            cfg,
		logger:         logger,
		store:          store,
		authService:    services.NewAuthService(store, cfg, pub),
		userService:    services.NewUserService(store),
		productService: services.NewProductService(store),
		uploadService:  services.NewUploadService(uploadProvider),
		cartService:    cartService,
		orderService:   services.NewOrderService(store, cartService),
	}, nil
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.New()

	// Add middlewares
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(s.corsMiddleware())

	// Setup routes
	router.GET("/health", s.healthCheckHandler)

	// Swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.StaticFile("/api-docs", "./docs/rapidoc.html")

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", s.registerHandler)
			auth.POST("/login", s.loginHandler)
			auth.POST("/refresh-token", s.refreshTokenHandler)
			auth.POST("/logout", s.logoutHandler)
		}

		protected := api.Group("/")

		protected.Use(s.AuthMiddleware())
		{
			user := protected.Group("/user")
			{
				user.GET("/profile", s.GetProfile)
				user.PUT("/profile", s.UpdateProfile)
			}

			// category routes
			categories := protected.Group("/categories")
			{
				categories.POST("", s.AdminAuthMiddleware(), s.CreateCategory)
				categories.PUT("/:id", s.AdminAuthMiddleware(), s.UpdateCategory)
				categories.DELETE("/:id", s.AdminAuthMiddleware(), s.DeleteCategory)
			}

			// product routes
			products := protected.Group("/products")
			{
				products.POST("", s.AdminAuthMiddleware(), s.CreateProduct)
				products.PUT("/:id", s.AdminAuthMiddleware(), s.UpdateProductByID)
				products.DELETE("/:id", s.AdminAuthMiddleware(), s.DeleteProductByID)
				products.POST("/:id/image", s.AdminAuthMiddleware(), s.UploadProductImage)
			}

			// cart routes
			cart := protected.Group("/cart")
			{
				cart.GET("", s.GetCart)
				cart.POST("/items", s.AddToCart)
				cart.PUT("/items/:itemId", s.UpdateCartItem)
				cart.DELETE("/items/:itemId", s.RemoveCartItem)
				cart.DELETE("", s.ClearCart)
			}

			// order routes
			orders := protected.Group("/orders")
			{
				orders.POST("", s.CreateOrder)
				orders.GET("", s.GetOrders)
				orders.GET("/:id", s.GetOrder)
				orders.POST("/:id/cancel", s.CancelOrder)
				orders.PUT("/:id/status", s.AdminAuthMiddleware(), s.UpdateOrderStatus)
			}
		}

		// public routes
		public := api.Group("/")
		{
			public.GET("/categories", s.GetCategories)
			public.GET("/products", s.GetProducts)
			public.GET("/products/:id", s.GetProductByID)
		}
	}

	return router
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (s *Server) healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
