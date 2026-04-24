package models

// PaginationParams contiene los parámetros de paginación parseados del request.
type PaginationParams struct {
	Page  int
	Limit int
}

// Offset calcula el offset SQL a partir de page y limit.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginationMeta es la metadata de paginación incluida en la respuesta.
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewPaginationMeta construye la metadata calculando total_pages.
func NewPaginationMeta(page, limit, total int) PaginationMeta {
	totalPages := 0
	if limit > 0 && total > 0 {
		totalPages = total / limit
		if total%limit != 0 {
			totalPages++
		}
	}
	return PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
