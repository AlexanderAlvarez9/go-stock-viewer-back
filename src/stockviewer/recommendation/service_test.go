package recommendation

import (
	"context"
	"testing"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
	"github.com/user/go-stock-viewer-back/src/stockviewer/mocks"
)

func TestGetTopRecommendations_Success(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	service := NewService(mockRepo)

	recommendations, err := service.GetTopRecommendations(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recommendations) == 0 {
		t.Fatal("expected recommendations, got empty slice")
	}

	for i := 1; i < len(recommendations); i++ {
		if recommendations[i].Score > recommendations[i-1].Score {
			t.Errorf("recommendations not sorted correctly: %v > %v at position %d",
				recommendations[i].Score, recommendations[i-1].Score, i)
		}
	}
}

func TestGetTopRecommendations_WithRanks(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	service := NewService(mockRepo)

	recommendations, err := service.GetTopRecommendations(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, rec := range recommendations {
		expectedRank := i + 1
		if rec.Rank != expectedRank {
			t.Errorf("expected rank %d, got %d", expectedRank, rec.Rank)
		}
	}
}

func TestGetTopRecommendations_LimitExceeds(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	service := NewService(mockRepo)

	recommendations, err := service.GetTopRecommendations(context.Background(), 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recommendations) > len(mockRepo.Stocks) {
		t.Errorf("expected at most %d recommendations, got %d",
			len(mockRepo.Stocks), len(recommendations))
	}
}

func TestCalculateScore(t *testing.T) {
	mockRepo := mocks.NewMockStocksRepository()
	service := NewService(mockRepo)

	tests := []struct {
		name     string
		stock    stockviewer.Stock
		minScore float64
		maxScore float64
	}{
		{
			name: "Strong buy with price increase",
			stock: stockviewer.Stock{
				RatingTo:   "Buy",
				Action:     "target raised by",
				TargetFrom: 100,
				TargetTo:   150,
			},
			minScore: 70,
			maxScore: 100,
		},
		{
			name: "Sell with price decrease",
			stock: stockviewer.Stock{
				RatingTo:   "Sell",
				Action:     "downgraded by",
				TargetFrom: 100,
				TargetTo:   50,
			},
			minScore: 0,
			maxScore: 30,
		},
		{
			name: "Neutral with no action",
			stock: stockviewer.Stock{
				RatingTo: "Neutral",
			},
			minScore: 30,
			maxScore: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.CalculateScore(tt.stock)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f",
					tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestGenerateReason(t *testing.T) {
	tests := []struct {
		name          string
		stock         stockviewer.Stock
		shouldContain string
	}{
		{
			name: "Buy rating",
			stock: stockviewer.Stock{
				RatingTo: "Buy",
			},
			shouldContain: "buy",
		},
		{
			name: "Target raised",
			stock: stockviewer.Stock{
				Action: "target raised by",
			},
			shouldContain: "increased",
		},
		{
			name: "Upgraded",
			stock: stockviewer.Stock{
				Action: "upgraded by",
			},
			shouldContain: "upgraded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := generateReason(tt.stock)
			if reason == "" {
				t.Error("expected non-empty reason")
			}
		})
	}
}
