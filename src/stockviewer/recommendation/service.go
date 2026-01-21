package recommendation

import (
	"context"
	"math"
	"sort"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
)

type Service struct {
	stocksRepo stockviewer.StocksRepository
}

func NewService(stocksRepo stockviewer.StocksRepository) *Service {
	return &Service{
		stocksRepo: stocksRepo,
	}
}

func (s *Service) GetTopRecommendations(ctx context.Context, limit int) ([]stockviewer.StockRecommendation, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}

	stocks, err := s.stocksRepo.GetTopRecommended(ctx, limit*2)
	if err != nil {
		return nil, err
	}

	var recommendations []stockviewer.StockRecommendation
	for _, stock := range stocks {
		rec := stockviewer.StockRecommendation{
			Stock:  stock,
			Score:  s.CalculateScore(stock),
			Reason: generateReason(stock),
		}
		recommendations = append(recommendations, rec)
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	for i := range recommendations {
		recommendations[i].Rank = i + 1
	}

	return recommendations, nil
}

func (s *Service) CalculateScore(stock stockviewer.Stock) float64 {
	score := 0.0

	ratingWeight := 0.40
	actionWeight := 0.35
	priceTargetWeight := 0.25

	ratingScore := calculateRatingScore(stock.RatingTo)
	score += ratingScore * ratingWeight

	actionScore := calculateActionScore(stock.Action)
	score += actionScore * actionWeight

	priceTargetScore := calculatePriceTargetScore(stock.TargetFrom, stock.TargetTo)
	score += priceTargetScore * priceTargetWeight

	normalizedScore := (score + 100) / 2
	return math.Round(normalizedScore*100) / 100
}

func calculateRatingScore(rating string) float64 {
	scores := map[string]float64{
		"Buy":            100.0,
		"Strong Buy":     100.0,
		"Outperform":     80.0,
		"Overweight":     70.0,
		"Accumulate":     60.0,
		"Hold":           40.0,
		"Neutral":        35.0,
		"Market Perform": 30.0,
		"Equal Weight":   30.0,
		"Underperform":   15.0,
		"Underweight":    15.0,
		"Reduce":         10.0,
		"Sell":           0.0,
		"Speculative":    50.0,
	}

	if score, ok := scores[rating]; ok {
		return score
	}
	return 40.0
}

func calculateActionScore(action string) float64 {
	scores := map[string]float64{
		"target raised by":  100.0,
		"upgraded by":       100.0,
		"initiated by":      60.0,
		"reiterated by":     50.0,
		"target lowered by": 0.0,
		"downgraded by":     0.0,
	}

	if score, ok := scores[action]; ok {
		return score
	}
	return 50.0
}

func calculatePriceTargetScore(from, to float64) float64 {
	if from <= 0 || to <= 0 {
		return 50.0
	}

	percentChange := ((to - from) / from) * 100

	if percentChange > 50 {
		return 100.0
	}
	if percentChange > 20 {
		return 80.0
	}
	if percentChange > 10 {
		return 70.0
	}
	if percentChange > 0 {
		return 60.0
	}
	if percentChange > -10 {
		return 40.0
	}
	if percentChange > -20 {
		return 20.0
	}
	return 0.0
}

func generateReason(stock stockviewer.Stock) string {
	var reasons []string

	switch stock.RatingTo {
	case "Buy", "Strong Buy":
		reasons = append(reasons, "Strong buy recommendation from analyst")
	case "Outperform", "Overweight":
		reasons = append(reasons, "Expected to outperform the market")
	case "Hold", "Neutral":
		reasons = append(reasons, "Stable performance expected")
	case "Sell", "Underperform":
		reasons = append(reasons, "Caution advised - underperformance expected")
	}

	switch stock.Action {
	case "target raised by":
		reasons = append(reasons, "Price target recently increased")
	case "upgraded by":
		reasons = append(reasons, "Recently upgraded by analyst")
	case "target lowered by":
		reasons = append(reasons, "Price target recently decreased")
	case "downgraded by":
		reasons = append(reasons, "Recently downgraded by analyst")
	}

	if stock.TargetFrom > 0 && stock.TargetTo > 0 {
		change := ((stock.TargetTo - stock.TargetFrom) / stock.TargetFrom) * 100
		if change > 10 {
			reasons = append(reasons, "Significant upside potential in price target")
		} else if change < -10 {
			reasons = append(reasons, "Notable downside risk in price target")
		}
	}

	if len(reasons) == 0 {
		return "Based on current market analysis"
	}

	result := reasons[0]
	for i := 1; i < len(reasons) && i < 3; i++ {
		result += ". " + reasons[i]
	}
	return result
}
