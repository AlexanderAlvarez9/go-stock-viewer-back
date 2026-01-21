package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/user/go-stock-viewer-back/src/stockviewer"
)

// Ping godoc
// @Summary      Health check endpoint
// @Description  Returns pong to verify the service is running
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  SuccessResponse
// @Router       /ping [get]
func (a *API) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data:    "pong",
		Message: "Service is running",
	})
}

// HealthCheck godoc
// @Summary      Detailed health check
// @Description  Returns detailed health status of the service
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  SuccessResponse
// @Router       /health [get]
func (a *API) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data: map[string]string{
			"status":  "healthy",
			"service": "go-stock-viewer-back",
		},
	})
}

// GetStocks godoc
// @Summary      List stocks
// @Description  Get a paginated list of stocks with optional filters
// @Tags         stocks
// @Accept       json
// @Produce      json
// @Param        ticker     query     string  false  "Filter by ticker symbol"
// @Param        company    query     string  false  "Filter by company name"
// @Param        brokerage  query     string  false  "Filter by brokerage"
// @Param        rating     query     string  false  "Filter by rating"
// @Param        action     query     string  false  "Filter by action"
// @Param        sort_by    query     string  false  "Sort by field (ticker, company, recommend_score, created_at)"
// @Param        sort_order query     string  false  "Sort order (ASC, DESC)"
// @Param        page       query     int     false  "Page number"  default(1)
// @Param        page_size  query     int     false  "Items per page"  default(20)
// @Success      200  {object}  PaginatedSuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/stocks [get]
func (a *API) GetStocks(c *gin.Context) {
	var filter stockviewer.StockFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid parameters",
			Message: err.Error(),
		})
		return
	}

	result, err := a.stocksService.GetStocks(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginatedSuccessResponse{
		Data:       result.Data,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: result.TotalPages,
	})
}

// GetStockByID godoc
// @Summary      Get stock by ID
// @Description  Get detailed information about a specific stock
// @Tags         stocks
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Stock ID"
// @Success      200  {object}  SuccessResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/stocks/{id} [get]
func (a *API) GetStockByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Message: "Stock ID is required",
		})
		return
	}

	stock, err := a.stocksService.GetStock(c.Request.Context(), id)
	if err != nil {
		if err == stockviewer.ErrStockNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not found",
				Message: "Stock not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: stock,
	})
}

// SearchStocks godoc
// @Summary      Search stocks
// @Description  Search stocks by ticker or company name
// @Tags         stocks
// @Accept       json
// @Produce      json
// @Param        q      query     string  true   "Search query"
// @Param        limit  query     int     false  "Maximum results"  default(10)
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/stocks/search [get]
func (a *API) SearchStocks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid query",
			Message: "Search query is required",
		})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	stocks, err := a.stocksService.SearchStocks(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: stocks,
	})
}

// GetFilters godoc
// @Summary      Get available filters
// @Description  Get available filter options for stocks (brokerages, ratings, actions)
// @Tags         stocks
// @Accept       json
// @Produce      json
// @Success      200  {object}  SuccessResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/stocks/filters [get]
func (a *API) GetFilters(c *gin.Context) {
	filters, err := a.stocksService.GetFilters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: filters,
	})
}

// GetRecommendations godoc
// @Summary      Get stock recommendations
// @Description  Get top recommended stocks based on the recommendation algorithm
// @Tags         recommendations
// @Accept       json
// @Produce      json
// @Param        limit  query     int     false  "Maximum recommendations"  default(10)
// @Success      200  {object}  SuccessResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/recommendations [get]
func (a *API) GetRecommendations(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	recommendations, err := a.recommendationService.GetTopRecommendations(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: recommendations,
	})
}

// SyncStocks godoc
// @Summary      Sync stocks from external API
// @Description  Fetch and synchronize stocks from the external KarenAI API
// @Tags         sync
// @Accept       json
// @Produce      json
// @Security     BasicAuth
// @Success      200  {object}  SyncResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse  "Sync already in progress"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/sync [post]
func (a *API) SyncStocks(c *gin.Context) {
	status, err := a.stocksService.SyncStocks(c.Request.Context())
	if err != nil {
		if err == stockviewer.ErrSyncInProgress {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: "Sync already in progress",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SyncResponse{
		Status:         status.Status,
		TotalRecords:   status.TotalRecords,
		NewRecords:     status.NewRecords,
		UpdatedRecords: status.UpdatedRecords,
		LastSync:       status.LastSync.Format("2006-01-02T15:04:05Z07:00"),
	})
}
