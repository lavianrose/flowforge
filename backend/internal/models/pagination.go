package models

import "math"

type PaginationParams struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	OrderBy  string `query:"order_by"`
	OrderDir string `query:"order_dir"` // asc, desc
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 20 // default
	}
	if p.OrderBy == "" {
		p.OrderBy = "created_at"
	}
	if p.OrderDir != "asc" && p.OrderDir != "desc" {
		p.OrderDir = "desc"
	}
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func CalculateTotalPages(total, perPage int) int {
	if perPage == 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(perPage)))
}

func NewPaginatedResponse(data interface{}, page, perPage, total int) *PaginatedResponse {
	return &PaginatedResponse{
		Data: data,
		Pagination: PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: CalculateTotalPages(total, perPage),
		},
	}
}
