package stockviewer

import (
	"context"
	"time"
)

type StockID string
type Rating string
type Action string

const (
	RatingBuy            Rating = "Buy"
	RatingNeutral        Rating = "Neutral"
	RatingMarketPerform  Rating = "Market Perform"
	RatingSell           Rating = "Sell"
	RatingSpeculative    Rating = "Speculative"
	RatingHold           Rating = "Hold"
	RatingOutperform     Rating = "Outperform"
	RatingUnderperform   Rating = "Underperform"
)

const (
	ActionTargetRaised  Action = "target raised by"
	ActionTargetLowered Action = "target lowered by"
	ActionUpgraded      Action = "upgraded by"
	ActionDowngraded    Action = "downgraded by"
	ActionInitiated     Action = "initiated by"
)

type Stock struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Ticker         string    `json:"ticker" gorm:"index;not null"`
	Company        string    `json:"company" gorm:"not null"`
	Brokerage      string    `json:"brokerage"`
	Action         string    `json:"action"`
	RatingFrom     string    `json:"rating_from"`
	RatingTo       string    `json:"rating_to"`
	TargetFrom     float64   `json:"target_from"`
	TargetTo       float64   `json:"target_to"`
	RecommendScore float64   `json:"recommend_score" gorm:"index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StockRecommendation struct {
	Stock          Stock   `json:"stock"`
	Score          float64 `json:"score"`
	Reason         string  `json:"reason"`
	Rank           int     `json:"rank"`
}

type SyncStatus struct {
	LastSync      time.Time `json:"last_sync"`
	TotalRecords  int       `json:"total_records"`
	NewRecords    int       `json:"new_records"`
	UpdatedRecords int      `json:"updated_records"`
	Status        string    `json:"status"`
}

type PaginatedResponse struct {
	Data       []Stock `json:"data"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalItems int64   `json:"total_items"`
	TotalPages int     `json:"total_pages"`
}

type StockFilter struct {
	Ticker    string `form:"ticker"`
	Company   string `form:"company"`
	Brokerage string `form:"brokerage"`
	Rating    string `form:"rating"`
	Action    string `form:"action"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

type StocksRepository interface {
	Save(ctx context.Context, stock Stock) error
	SaveBatch(ctx context.Context, stocks []Stock) error
	GetByID(ctx context.Context, id string) (*Stock, error)
	GetByTicker(ctx context.Context, ticker string) ([]Stock, error)
	GetAll(ctx context.Context, filter StockFilter) ([]Stock, int64, error)
	GetTopRecommended(ctx context.Context, limit int) ([]Stock, error)
	Search(ctx context.Context, query string, limit int) ([]Stock, error)
	Delete(ctx context.Context, id string) error
	GetDistinctBrokerages(ctx context.Context) ([]string, error)
	GetDistinctRatings(ctx context.Context) ([]string, error)
}

type StocksFetcher interface {
	FetchStocks(ctx context.Context) (<-chan StockOrError, error)
}

type StocksService interface {
	SyncStocks(ctx context.Context) (*SyncStatus, error)
	GetStock(ctx context.Context, id string) (*Stock, error)
	GetStocks(ctx context.Context, filter StockFilter) (*PaginatedResponse, error)
	SearchStocks(ctx context.Context, query string, limit int) ([]Stock, error)
	GetFilters(ctx context.Context) (*FiltersResponse, error)
}

type RecommendationService interface {
	GetTopRecommendations(ctx context.Context, limit int) ([]StockRecommendation, error)
	CalculateScore(stock Stock) float64
}

type StockOrError struct {
	Stock Stock
	Error error
}

type FiltersResponse struct {
	Brokerages []string `json:"brokerages"`
	Ratings    []string `json:"ratings"`
	Actions    []string `json:"actions"`
}
