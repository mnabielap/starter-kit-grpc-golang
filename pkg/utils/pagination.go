package utils

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type PaginationScope struct {
	Page  int32
	Limit int32
	Sort  string
}

// Paginate returns a GORM scope for pagination
func (p *PaginationScope) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page := p.Page
		if page < 1 {
			page = 1
		}

		limit := p.Limit
		if limit < 1 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}

		offset := (page - 1) * limit
		return db.Offset(int(offset)).Limit(int(limit))
	}
}

// SortScope returns a GORM scope for sorting safely
// allowedFields is a map of "Input Field Name" -> "DB Column Name"
func (p *PaginationScope) SortScope(allowedFields map[string]string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		sortParam := p.Sort
		if sortParam == "" {
			return db.Order("created_at desc") // Default
		}

		// Expected format: "field:direction" (e.g., "name:asc")
		parts := strings.Split(sortParam, ":")
		rawField := parts[0]
		direction := "asc"
		if len(parts) > 1 && strings.ToLower(parts[1]) == "desc" {
			direction = "desc"
		}

		// Map input field to DB column
		dbColumn, valid := allowedFields[rawField]
		if valid {
			return db.Order(fmt.Sprintf("%s %s", dbColumn, direction))
		}

		// Fallback
		return db.Order("created_at desc")
	}
}