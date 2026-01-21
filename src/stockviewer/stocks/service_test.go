package stocks

import (
	"context"
	"errors"
	"testing"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
	"github.com/user/go-stock-viewer-back/src/stockviewer/mocks"
)

func TestGetStocks_Success(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	mockFetcher := mocks.NewMockStocksFetcher()
	service := NewService(mockRepo, mockFetcher)

	filter := stockviewer.StockFilter{
		Page:     1,
		PageSize: 10,
	}

	result, err := service.GetStocks(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.TotalItems != int64(len(mockRepo.Stocks)) {
		t.Errorf("expected %d total items, got %d", len(mockRepo.Stocks), result.TotalItems)
	}
}

func TestGetStocks_WithPagination(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	mockFetcher := mocks.NewMockStocksFetcher()
	service := NewService(mockRepo, mockFetcher)

	filter := stockviewer.StockFilter{
		Page:     1,
		PageSize: 2,
	}

	result, err := service.GetStocks(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}

	if result.PageSize != 2 {
		t.Errorf("expected page size 2, got %d", result.PageSize)
	}
}

func TestGetStock_Success(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	mockFetcher := mocks.NewMockStocksFetcher()
	service := NewService(mockRepo, mockFetcher)

	stock, err := service.GetStock(context.Background(), "test-id-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stock == nil {
		t.Fatal("expected stock, got nil")
	}

	if stock.Ticker != "AAPL" {
		t.Errorf("expected ticker AAPL, got %s", stock.Ticker)
	}
}

func TestGetStock_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	mockFetcher := mocks.NewMockStocksFetcher()
	service := NewService(mockRepo, mockFetcher)

	_, err := service.GetStock(context.Background(), "non-existent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, stockviewer.ErrStockNotFound) {
		t.Errorf("expected ErrStockNotFound, got %v", err)
	}
}

func TestSearchStocks_Success(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	mockFetcher := mocks.NewMockStocksFetcher()
	service := NewService(mockRepo, mockFetcher)

	stocks, err := service.SearchStocks(context.Background(), "AAPL", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stocks == nil {
		t.Fatal("expected stocks, got nil")
	}
}

func TestSyncStocks_Success(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	mockFetcher := mocks.NewMockStocksFetcher()
	service := NewService(mockRepo, mockFetcher)

	status, err := service.SyncStocks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status == nil {
		t.Fatal("expected status, got nil")
	}

	if status.Status != "completed" {
		t.Errorf("expected status completed, got %s", status.Status)
	}
}

func TestSyncStocks_AlreadyInProgress(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	mockFetcher := &slowMockFetcher{}
	service := NewService(mockRepo, mockFetcher)

	go func() {
		service.SyncStocks(context.Background())
	}()

	for !service.syncInProg {
	}

	_, err := service.SyncStocks(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, stockviewer.ErrSyncInProgress) {
		t.Errorf("expected ErrSyncInProgress, got %v", err)
	}
}

type slowMockFetcher struct{}

func (m *slowMockFetcher) FetchStocks(ctx context.Context) (<-chan stockviewer.StockOrError, error) {
	ch := make(chan stockviewer.StockOrError)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
			return
		}
	}()
	return ch, nil
}

func TestCalculateRecommendScore(t *testing.T) {
	tests := []struct {
		name     string
		stock    stockviewer.Stock
		minScore float64
		maxScore float64
	}{
		{
			name: "Buy rating with target raised",
			stock: stockviewer.Stock{
				RatingTo: "Buy",
				Action:   "target raised by",
			},
			minScore: 70,
			maxScore: 100,
		},
		{
			name: "Sell rating with target lowered",
			stock: stockviewer.Stock{
				RatingTo: "Sell",
				Action:   "target lowered by",
			},
			minScore: 0,
			maxScore: 30,
		},
		{
			name: "Neutral rating",
			stock: stockviewer.Stock{
				RatingTo: "Neutral",
				Action:   "initiated by",
			},
			minScore: 40,
			maxScore: 70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateRecommendScore(tt.stock)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}
