package data

type Paginator struct {
	Page     int64
	PageSize int64
}

type Metadata struct {
	CurrentPage  int64 `json:"current_page,omitempty"`
	PageSize     int64 `json:"page_size,omitempty"`
	FirstPage    int64 `json:"first_page,omitempty"`
	LastPage     int64 `json:"last_page,omitempty"`
	TotalRecords int64 `json:"total_records,omitempty"`
}

func (p Paginator) limit() int64 {
	return p.PageSize
}

func (p Paginator) offset() int64 {
	return (p.Page - 1) * p.PageSize
}

// calculateMetadata returns metadata regarding pagination.
func calculateMetadata(totalRecords, page, pageSize int64) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}
