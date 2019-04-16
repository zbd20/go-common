package models

type Page struct {
	PageSize   int64  `json:"limit"`
	Offset     int64  `json:"offset"`
	Page       int64  `json:"page"`
	TotalCount int64  `json:"-"`
	Query      string `json:"-"`
	OrderBy    string `json:"order_by"`
	Sort       string `json:"sort"`
}

type Result struct {
	TotalCount  *int64      `json:"total_count,omitempty"`
	PageCount   *int64      `json:"page_count,omitempty"`
	CurrentPage *int64      `json:"current_page,omitempty"`
	PageSize    *int64      `json:"page_size,omitempty"`
	Results     interface{} `json:"result"`
	Code        int64       `json:"code"`
}

func NewResult(count int64, page *Page, results interface{}) Result {
	var result Result
	var pageCount int64
	if page != nil {
		result = Result{
			TotalCount:  &count,
			CurrentPage: &page.Page,
			PageSize:    &page.PageSize,
			PageCount:   &pageCount,
			Results:     results,
			Code:        0,
		}

		pc := count / page.PageSize
		result.PageCount = &pc
		if count%page.PageSize > 0 {
			*(result.PageCount) += 1
		}
	} else {
		result = Result{
			Results: results,
			Code:    0,
		}
	}

	return result
}
