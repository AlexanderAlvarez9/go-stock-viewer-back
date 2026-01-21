package mocks

import (
	"context"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
)

type MockStocksFetcher struct {
	Stocks []stockviewer.Stock
	Error  error
}

func NewMockStocksFetcher() *MockStocksFetcher {
	return &MockStocksFetcher{
		Stocks: []stockviewer.Stock{
			{
				ID:         "mock-1",
				Ticker:     "RMTI",
				Company:    "Rockwell Medical",
				Brokerage:  "Analyst Firm",
				Action:     "target lowered by",
				RatingTo:   "Buy",
			},
			{
				ID:         "mock-2",
				Ticker:     "AKBA",
				Company:    "Akebia Therapeutics",
				Brokerage:  "Analyst Firm",
				Action:     "target lowered by",
				RatingTo:   "Buy",
			},
			{
				ID:         "mock-3",
				Ticker:     "CECO",
				Company:    "CECO Environmental",
				Brokerage:  "Analyst Firm",
				Action:     "target raised by",
				RatingTo:   "Buy",
			},
		},
	}
}

func (m *MockStocksFetcher) FetchStocks(ctx context.Context) (<-chan stockviewer.StockOrError, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	ch := make(chan stockviewer.StockOrError, len(m.Stocks))

	go func() {
		defer close(ch)
		for _, stock := range m.Stocks {
			select {
			case <-ctx.Done():
				ch <- stockviewer.StockOrError{Error: ctx.Err()}
				return
			case ch <- stockviewer.StockOrError{Stock: stock}:
			}
		}
	}()

	return ch, nil
}
