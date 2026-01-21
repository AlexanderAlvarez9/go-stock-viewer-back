package stocks

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) (*Storage, error) {
	if err := db.AutoMigrate(&stockviewer.Stock{}); err != nil {
		return nil, stockviewer.StorageError{Operation: "migrate", Err: err}
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, stock stockviewer.Stock) error {
	result := s.db.WithContext(ctx).Save(&stock)
	if result.Error != nil {
		return stockviewer.StorageError{Operation: "save", Err: result.Error}
	}
	return nil
}

func (s *Storage) SaveBatch(ctx context.Context, stocks []stockviewer.Stock) error {
	if len(stocks) == 0 {
		return nil
	}

	result := s.db.WithContext(ctx).Save(&stocks)
	if result.Error != nil {
		return stockviewer.StorageError{Operation: "save_batch", Err: result.Error}
	}
	return nil
}

func (s *Storage) GetByID(ctx context.Context, id string) (*stockviewer.Stock, error) {
	var stock stockviewer.Stock
	result := s.db.WithContext(ctx).Where("id = ?", id).First(&stock)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, stockviewer.ErrStockNotFound
		}
		return nil, stockviewer.StorageError{Operation: "get_by_id", Err: result.Error}
	}
	return &stock, nil
}

func (s *Storage) GetByTicker(ctx context.Context, ticker string) ([]stockviewer.Stock, error) {
	var stocks []stockviewer.Stock
	result := s.db.WithContext(ctx).Where("ticker = ?", ticker).Find(&stocks)
	if result.Error != nil {
		return nil, stockviewer.StorageError{Operation: "get_by_ticker", Err: result.Error}
	}
	return stocks, nil
}

func (s *Storage) GetAll(ctx context.Context, filter stockviewer.StockFilter) ([]stockviewer.Stock, int64, error) {
	var stocks []stockviewer.Stock
	var total int64

	query := s.db.WithContext(ctx).Model(&stockviewer.Stock{})

	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, stockviewer.StorageError{Operation: "count", Err: err}
	}

	query = applySorting(query, filter)
	query = applyPagination(query, filter)

	if err := query.Find(&stocks).Error; err != nil {
		return nil, 0, stockviewer.StorageError{Operation: "get_all", Err: err}
	}

	return stocks, total, nil
}

func (s *Storage) GetTopRecommended(ctx context.Context, limit int) ([]stockviewer.Stock, error) {
	var stocks []stockviewer.Stock
	result := s.db.WithContext(ctx).
		Order("recommend_score DESC").
		Limit(limit).
		Find(&stocks)

	if result.Error != nil {
		return nil, stockviewer.StorageError{Operation: "get_top_recommended", Err: result.Error}
	}
	return stocks, nil
}

func (s *Storage) Search(ctx context.Context, query string, limit int) ([]stockviewer.Stock, error) {
	var stocks []stockviewer.Stock
	searchPattern := fmt.Sprintf("%%%s%%", strings.ToLower(query))

	result := s.db.WithContext(ctx).
		Where("LOWER(ticker) LIKE ? OR LOWER(company) LIKE ?", searchPattern, searchPattern).
		Order("recommend_score DESC").
		Limit(limit).
		Find(&stocks)

	if result.Error != nil {
		return nil, stockviewer.StorageError{Operation: "search", Err: result.Error}
	}
	return stocks, nil
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&stockviewer.Stock{}, "id = ?", id)
	if result.Error != nil {
		return stockviewer.StorageError{Operation: "delete", Err: result.Error}
	}
	if result.RowsAffected == 0 {
		return stockviewer.ErrStockNotFound
	}
	return nil
}

func (s *Storage) GetDistinctBrokerages(ctx context.Context) ([]string, error) {
	var brokerages []string
	result := s.db.WithContext(ctx).
		Model(&stockviewer.Stock{}).
		Distinct("brokerage").
		Where("brokerage != ''").
		Pluck("brokerage", &brokerages)

	if result.Error != nil {
		return nil, stockviewer.StorageError{Operation: "get_distinct_brokerages", Err: result.Error}
	}
	return brokerages, nil
}

func (s *Storage) GetDistinctRatings(ctx context.Context) ([]string, error) {
	var ratings []string
	result := s.db.WithContext(ctx).
		Model(&stockviewer.Stock{}).
		Distinct("rating_to").
		Where("rating_to != ''").
		Pluck("rating_to", &ratings)

	if result.Error != nil {
		return nil, stockviewer.StorageError{Operation: "get_distinct_ratings", Err: result.Error}
	}
	return ratings, nil
}

func applyFilters(query *gorm.DB, filter stockviewer.StockFilter) *gorm.DB {
	if filter.Ticker != "" {
		query = query.Where("LOWER(ticker) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Ticker)))
	}
	if filter.Company != "" {
		query = query.Where("LOWER(company) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Company)))
	}
	if filter.Brokerage != "" {
		query = query.Where("brokerage = ?", filter.Brokerage)
	}
	if filter.Rating != "" {
		query = query.Where("rating_to = ?", filter.Rating)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	return query
}

func applySorting(query *gorm.DB, filter stockviewer.StockFilter) *gorm.DB {
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "recommend_score"
	}

	validSortFields := map[string]bool{
		"ticker":          true,
		"company":         true,
		"brokerage":       true,
		"recommend_score": true,
		"created_at":      true,
		"updated_at":      true,
	}

	if !validSortFields[sortBy] {
		sortBy = "recommend_score"
	}

	sortOrder := strings.ToUpper(filter.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	return query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
}

func applyPagination(query *gorm.DB, filter stockviewer.StockFilter) *gorm.DB {
	page := filter.Page
	if page < 1 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return query.Offset(offset).Limit(pageSize)
}
