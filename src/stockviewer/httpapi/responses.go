package httpapi

import "github.com/user/go-stock-viewer-back/src/stockviewer"

type SuccessResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

type PaginatedSuccessResponse struct {
	Data       []stockviewer.Stock `json:"data"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalItems int64                `json:"total_items"`
	TotalPages int                  `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SyncResponse struct {
	Status         string `json:"status"`
	TotalRecords   int    `json:"total_records"`
	NewRecords     int    `json:"new_records"`
	UpdatedRecords int    `json:"updated_records"`
	LastSync       string `json:"last_sync"`
}

type FiltersResponse struct {
	Brokerages []string `json:"brokerages"`
	Ratings    []string `json:"ratings"`
	Actions    []string `json:"actions"`
}
