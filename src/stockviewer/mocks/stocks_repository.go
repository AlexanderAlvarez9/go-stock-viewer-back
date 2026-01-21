package mocks

import (
	"context"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
)

type MockStocksRepository struct {
	Stocks     []stockviewer.Stock
	Error      error
	SaveError  error
}

func NewMockStocksRepository() *MockStocksRepository {
	return &MockStocksRepository{
		Stocks: []stockviewer.Stock{
			{
				ID:             "test-id-1",
				Ticker:         "AAPL",
				Company:        "Apple Inc.",
				Brokerage:      "Goldman Sachs",
				Action:         "target raised by",
				RatingFrom:     "Hold",
				RatingTo:       "Buy",
				TargetFrom:     150.0,
				TargetTo:       180.0,
				RecommendScore: 85.5,
			},
			{
				ID:             "test-id-2",
				Ticker:         "GOOGL",
				Company:        "Alphabet Inc.",
				Brokerage:      "Morgan Stanley",
				Action:         "upgraded by",
				RatingFrom:     "Neutral",
				RatingTo:       "Buy",
				TargetFrom:     2800.0,
				TargetTo:       3200.0,
				RecommendScore: 90.0,
			},
			{
				ID:             "test-id-3",
				Ticker:         "MSFT",
				Company:        "Microsoft Corporation",
				Brokerage:      "JP Morgan",
				Action:         "target lowered by",
				RatingFrom:     "Buy",
				RatingTo:       "Neutral",
				TargetFrom:     350.0,
				TargetTo:       320.0,
				RecommendScore: 45.0,
			},
		},
	}
}

func (m *MockStocksRepository) Save(ctx context.Context, stock stockviewer.Stock) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.Stocks = append(m.Stocks, stock)
	return nil
}

func (m *MockStocksRepository) SaveBatch(ctx context.Context, stocks []stockviewer.Stock) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.Stocks = append(m.Stocks, stocks...)
	return nil
}

func (m *MockStocksRepository) GetByID(ctx context.Context, id string) (*stockviewer.Stock, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	for _, stock := range m.Stocks {
		if stock.ID == id {
			return &stock, nil
		}
	}
	return nil, stockviewer.ErrStockNotFound
}

func (m *MockStocksRepository) GetByTicker(ctx context.Context, ticker string) ([]stockviewer.Stock, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	var result []stockviewer.Stock
	for _, stock := range m.Stocks {
		if stock.Ticker == ticker {
			result = append(result, stock)
		}
	}
	return result, nil
}

func (m *MockStocksRepository) GetAll(ctx context.Context, filter stockviewer.StockFilter) ([]stockviewer.Stock, int64, error) {
	if m.Error != nil {
		return nil, 0, m.Error
	}
	return m.Stocks, int64(len(m.Stocks)), nil
}

func (m *MockStocksRepository) GetTopRecommended(ctx context.Context, limit int) ([]stockviewer.Stock, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	if limit > len(m.Stocks) {
		limit = len(m.Stocks)
	}
	return m.Stocks[:limit], nil
}

func (m *MockStocksRepository) Search(ctx context.Context, query string, limit int) ([]stockviewer.Stock, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Stocks, nil
}

func (m *MockStocksRepository) Delete(ctx context.Context, id string) error {
	if m.Error != nil {
		return m.Error
	}
	for i, stock := range m.Stocks {
		if stock.ID == id {
			m.Stocks = append(m.Stocks[:i], m.Stocks[i+1:]...)
			return nil
		}
	}
	return stockviewer.ErrStockNotFound
}

func (m *MockStocksRepository) GetDistinctBrokerages(ctx context.Context) ([]string, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	brokerages := make(map[string]bool)
	for _, stock := range m.Stocks {
		if stock.Brokerage != "" {
			brokerages[stock.Brokerage] = true
		}
	}
	result := make([]string, 0, len(brokerages))
	for b := range brokerages {
		result = append(result, b)
	}
	return result, nil
}

func (m *MockStocksRepository) GetDistinctRatings(ctx context.Context) ([]string, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	ratings := make(map[string]bool)
	for _, stock := range m.Stocks {
		if stock.RatingTo != "" {
			ratings[stock.RatingTo] = true
		}
	}
	result := make([]string, 0, len(ratings))
	for r := range ratings {
		result = append(result, r)
	}
	return result, nil
}
