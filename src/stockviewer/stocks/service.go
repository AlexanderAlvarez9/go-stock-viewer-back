package stocks

import (
	"context"
	"log"
	"math"
	"sync"
	"time"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
)

type Service struct {
	storage     stockviewer.StocksRepository
	fetcher     stockviewer.StocksFetcher
	syncMutex   sync.Mutex
	syncInProg  bool
	lastSync    time.Time
}

func NewService(storage stockviewer.StocksRepository, fetcher stockviewer.StocksFetcher) *Service {
	return &Service{
		storage: storage,
		fetcher: fetcher,
	}
}

func (s *Service) SyncStocks(ctx context.Context) (*stockviewer.SyncStatus, error) {
	s.syncMutex.Lock()
	if s.syncInProg {
		s.syncMutex.Unlock()
		return nil, stockviewer.ErrSyncInProgress
	}
	s.syncInProg = true
	s.syncMutex.Unlock()

	defer func() {
		s.syncMutex.Lock()
		s.syncInProg = false
		s.syncMutex.Unlock()
	}()

	status := &stockviewer.SyncStatus{
		Status: "in_progress",
	}

	stocksChan, err := s.fetcher.FetchStocks(ctx)
	if err != nil {
		status.Status = "error"
		return status, err
	}

	var batch []stockviewer.Stock
	batchSize := 100
	totalRecords := 0
	newRecords := 0

	for stockOrErr := range stocksChan {
		if stockOrErr.Error != nil {
			log.Printf("Error fetching stock: %v", stockOrErr.Error)
			continue
		}

		stock := stockOrErr.Stock
		stock.RecommendScore = calculateRecommendScore(stock)
		stock.UpdatedAt = time.Now()

		existing, err := s.storage.GetByID(ctx, stock.ID)
		if err == stockviewer.ErrStockNotFound {
			stock.CreatedAt = time.Now()
			newRecords++
		} else if err == nil {
			stock.CreatedAt = existing.CreatedAt
		}

		batch = append(batch, stock)
		totalRecords++

		if len(batch) >= batchSize {
			if err := s.storage.SaveBatch(ctx, batch); err != nil {
				log.Printf("Error saving batch: %v", err)
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := s.storage.SaveBatch(ctx, batch); err != nil {
			log.Printf("Error saving final batch: %v", err)
		}
	}

	s.lastSync = time.Now()
	status.LastSync = s.lastSync
	status.TotalRecords = totalRecords
	status.NewRecords = newRecords
	status.UpdatedRecords = totalRecords - newRecords
	status.Status = "completed"

	return status, nil
}

func (s *Service) GetStock(ctx context.Context, id string) (*stockviewer.Stock, error) {
	return s.storage.GetByID(ctx, id)
}

func (s *Service) GetStocks(ctx context.Context, filter stockviewer.StockFilter) (*stockviewer.PaginatedResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	stocks, total, err := s.storage.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	return &stockviewer.PaginatedResponse{
		Data:       stocks,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (s *Service) SearchStocks(ctx context.Context, query string, limit int) ([]stockviewer.Stock, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.storage.Search(ctx, query, limit)
}

func (s *Service) GetFilters(ctx context.Context) (*stockviewer.FiltersResponse, error) {
	brokerages, err := s.storage.GetDistinctBrokerages(ctx)
	if err != nil {
		return nil, err
	}

	ratings, err := s.storage.GetDistinctRatings(ctx)
	if err != nil {
		return nil, err
	}

	actions := []string{
		string(stockviewer.ActionTargetRaised),
		string(stockviewer.ActionTargetLowered),
		string(stockviewer.ActionUpgraded),
		string(stockviewer.ActionDowngraded),
		string(stockviewer.ActionInitiated),
	}

	return &stockviewer.FiltersResponse{
		Brokerages: brokerages,
		Ratings:    ratings,
		Actions:    actions,
	}, nil
}

func calculateRecommendScore(stock stockviewer.Stock) float64 {
	score := 50.0

	ratingScores := map[string]float64{
		"Buy":            30.0,
		"Outperform":     25.0,
		"Overweight":     20.0,
		"Hold":           0.0,
		"Neutral":        -5.0,
		"Market Perform": -10.0,
		"Underperform":   -20.0,
		"Underweight":    -20.0,
		"Sell":           -30.0,
		"Speculative":    10.0,
	}

	if ratingScore, ok := ratingScores[stock.RatingTo]; ok {
		score += ratingScore
	}

	actionScores := map[string]float64{
		"target raised by": 15.0,
		"upgraded by":      20.0,
		"initiated by":     5.0,
		"target lowered by": -15.0,
		"downgraded by":    -20.0,
	}

	if actionScore, ok := actionScores[stock.Action]; ok {
		score += actionScore
	}

	if stock.TargetFrom > 0 && stock.TargetTo > 0 {
		priceChange := ((stock.TargetTo - stock.TargetFrom) / stock.TargetFrom) * 100
		score += priceChange * 0.5
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return math.Round(score*100) / 100
}
