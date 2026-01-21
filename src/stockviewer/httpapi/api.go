package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/user/go-stock-viewer-back/src/stockviewer"
)

type Config struct {
	StocksService         stockviewer.StocksService
	RecommendationService stockviewer.RecommendationService
	BasicAuthUser         string
	BasicAuthPassword     string
}

type API struct {
	stocksService         stockviewer.StocksService
	recommendationService stockviewer.RecommendationService
	basicAuthUser         string
	basicAuthPassword     string
}

func New(cfg Config) *API {
	return &API{
		stocksService:         cfg.StocksService,
		recommendationService: cfg.RecommendationService,
		basicAuthUser:         cfg.BasicAuthUser,
		basicAuthPassword:     cfg.BasicAuthPassword,
	}
}

func (a *API) ConfigureRoutes(router *gin.Engine) {
	router.Use(CORSMiddleware())

	router.GET("/ping", a.Ping)
	router.GET("/health", a.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/stocks", a.GetStocks)
		v1.GET("/stocks/search", a.SearchStocks)
		v1.GET("/stocks/:id", a.GetStockByID)
		v1.GET("/stocks/filters", a.GetFilters)

		v1.GET("/recommendations", a.GetRecommendations)

		protected := v1.Group("")
		protected.Use(a.BasicAuthMiddleware())
		{
			protected.POST("/sync", a.SyncStocks)
		}
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (a *API) BasicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth || user != a.basicAuthUser || password != a.basicAuthPassword {
			c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
			c.JSON(401, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid credentials",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
