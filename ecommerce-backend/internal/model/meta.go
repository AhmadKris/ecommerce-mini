package model

// Meta is the pagination envelope every list endpoint responds with, per
// the API convention in .claude/CLAUDE.md: {"items": [...], "meta": {...}}.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewMeta computes TotalPages from total/limit.
func NewMeta(page, limit int, total int64) Meta {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	return Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}
