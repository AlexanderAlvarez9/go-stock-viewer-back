package karenai

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/user/go-stock-viewer-back/src/stockviewer"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type APIResponse struct {
	Items    []StockItem `json:"items"`
	NextPage string      `json:"next_page"`
}

type StockItem struct {
	Ticker     string      `json:"ticker"`
	Company    string      `json:"company"`
	Brokerage  string      `json:"brokerage"`
	Action     string      `json:"action"`
	RatingFrom string      `json:"rating_from"`
	RatingTo   string      `json:"rating_to"`
	TargetFrom interface{} `json:"target_from"`
	TargetTo   interface{} `json:"target_to"`
}

func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case int:
		return float64(val)
	case int64:
		return float64(val)
	}
	return 0
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) FetchStocks(ctx context.Context) (<-chan stockviewer.StockOrError, error) {
	stocksChan := make(chan stockviewer.StockOrError, 100)

	go func() {
		defer close(stocksChan)

		nextPage := ""
		pageCount := 0
		maxPages := 100

		for pageCount < maxPages {
			select {
			case <-ctx.Done():
				stocksChan <- stockviewer.StockOrError{Error: ctx.Err()}
				return
			default:
			}

			response, err := c.fetchPage(ctx, nextPage)
			if err != nil {
				stocksChan <- stockviewer.StockOrError{Error: err}
				return
			}

			for _, item := range response.Items {
				stock := convertToStock(item)
				stocksChan <- stockviewer.StockOrError{Stock: stock}
			}

			if response.NextPage == "" {
				break
			}

			nextPage = response.NextPage
			pageCount++
		}
	}()

	return stocksChan, nil
}

func (c *Client) fetchPage(ctx context.Context, nextPage string) (*APIResponse, error) {
	url := fmt.Sprintf("%s/swechallenge/list", c.baseURL)
	if nextPage != "" {
		url = fmt.Sprintf("%s?next_page=%s", url, nextPage)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, stockviewer.ExternalAPIError{
			Service: "karenai",
			Message: fmt.Sprintf("error creating request: %v", err),
			Err:     err,
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, stockviewer.ExternalAPIError{
			Service: "karenai",
			Message: fmt.Sprintf("error making request: %v", err),
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, stockviewer.ExternalAPIError{
			Service:    "karenai",
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code: %s", string(body)),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, stockviewer.ExternalAPIError{
			Service: "karenai",
			Message: fmt.Sprintf("error reading response: %v", err),
			Err:     err,
		}
	}

	var response APIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, stockviewer.ExternalAPIError{
			Service: "karenai",
			Message: fmt.Sprintf("error parsing response: %v", err),
			Err:     err,
		}
	}

	return &response, nil
}

func convertToStock(item StockItem) stockviewer.Stock {
	targetFrom := parseFloat(item.TargetFrom)
	targetTo := parseFloat(item.TargetTo)
	id := generateStockID(item, targetFrom, targetTo)

	return stockviewer.Stock{
		ID:         id,
		Ticker:     item.Ticker,
		Company:    item.Company,
		Brokerage:  item.Brokerage,
		Action:     item.Action,
		RatingFrom: item.RatingFrom,
		RatingTo:   item.RatingTo,
		TargetFrom: targetFrom,
		TargetTo:   targetTo,
	}
}

func generateStockID(item StockItem, targetFrom, targetTo float64) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%.2f|%.2f",
		item.Ticker,
		item.Company,
		item.Brokerage,
		item.Action,
		item.RatingFrom,
		item.RatingTo,
		targetFrom,
		targetTo,
	)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
