package repository

import "strings"

// Pagination and sort-direction defaults shared by every admin list
// endpoint.
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MinPageSize     = 1
	MaxPageSize     = 100

	SortAsc  = "asc"
	SortDesc = "desc"
)

// ListParams captures the page/size/sort/search inputs every admin list
// endpoint accepts from its query string. Call Normalize before using the
// values to build a query.
type ListParams struct {
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
	Search   string
}

// Normalize clamps Page to >= 1 and PageSize to [MinPageSize, MaxPageSize],
// defaulting unset values to DefaultPage / DefaultPageSize, and trims
// whitespace from Search. It never returns an error: out-of-range input is
// clamped, not rejected.
func (p *ListParams) Normalize() {
	if p.Page < DefaultPage {
		p.Page = DefaultPage
	}

	if p.PageSize == 0 {
		p.PageSize = DefaultPageSize
	}
	switch {
	case p.PageSize < MinPageSize:
		p.PageSize = MinPageSize
	case p.PageSize > MaxPageSize:
		p.PageSize = MaxPageSize
	}

	p.Search = strings.TrimSpace(p.Search)
}

// Offset returns the SQL OFFSET for the current, normalized page.
func (p ListParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// ResolveSort maps a caller-supplied sort_by value to its allowlisted SQL
// column literal. allow must carry a "" entry naming the default column: an
// unrecognized (or empty) by silently falls back to that default rather than
// rejecting the request (decisions.md D6), keeping this the one path every
// later phase reuses for ORDER BY. The returned column is always the
// allowlist's own literal — never the caller's string — which is what makes
// this the single defence against ORDER BY injection (column names cannot be
// parameterized the way values can).
//
// dir is validated independently of the allowlist: any value other than a
// case-insensitive "desc" resolves to "asc".
func ResolveSort(allow map[string]string, by, dir string) (col, sortDir string) {
	col, ok := allow[by]
	if !ok {
		col = allow[""]
	}

	sortDir = SortAsc
	if strings.EqualFold(dir, SortDesc) {
		sortDir = SortDesc
	}

	return col, sortDir
}

// Paginated is the envelope every list endpoint returns.
type Paginated[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
