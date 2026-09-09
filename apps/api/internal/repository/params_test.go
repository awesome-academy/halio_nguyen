package repository

import "testing"

func TestListParamsNormalize(t *testing.T) {
	tests := []struct {
		name         string
		in           ListParams
		wantPage     int
		wantPageSize int
	}{
		{
			name:         "zero values default",
			in:           ListParams{},
			wantPage:     1,
			wantPageSize: 20,
		},
		{
			name:         "negative page clamps to 1",
			in:           ListParams{Page: -5, PageSize: 20},
			wantPage:     1,
			wantPageSize: 20,
		},
		{
			name:         "page size above max clamps to 100",
			in:           ListParams{Page: 1, PageSize: 500},
			wantPage:     1,
			wantPageSize: 100,
		},
		{
			name:         "negative page size clamps to min",
			in:           ListParams{Page: 1, PageSize: -10},
			wantPage:     1,
			wantPageSize: 1,
		},
		{
			name:         "valid values pass through unchanged",
			in:           ListParams{Page: 3, PageSize: 50},
			wantPage:     3,
			wantPageSize: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.in
			p.Normalize()
			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.wantPageSize)
			}
		})
	}
}

func TestListParamsOffset(t *testing.T) {
	p := ListParams{Page: 3, PageSize: 20}
	p.Normalize()
	if got, want := p.Offset(), 40; got != want {
		t.Errorf("Offset() = %d, want %d", got, want)
	}
}

func TestResolveSort(t *testing.T) {
	allow := map[string]string{
		"":           "created_at",
		"name":       "c.name",
		"created_at": "c.created_at",
	}

	tests := []struct {
		name    string
		by      string
		dir     string
		wantCol string
		wantDir string
	}{
		{
			name:    "allowlisted column with desc",
			by:      "name",
			dir:     "desc",
			wantCol: "c.name",
			wantDir: "desc",
		},
		{
			name:    "allowlisted column with uppercase dir",
			by:      "created_at",
			dir:     "DESC",
			wantCol: "c.created_at",
			wantDir: "desc",
		},
		{
			name:    "unknown column falls back to default, never echoed",
			by:      "unknown_column",
			dir:     "asc",
			wantCol: "created_at",
			wantDir: "asc",
		},
		{
			name:    "sql injection attempt falls back to default",
			by:      "id); DROP TABLE users;--",
			dir:     "asc",
			wantCol: "created_at",
			wantDir: "asc",
		},
		{
			name:    "empty by falls back to default",
			by:      "",
			dir:     "",
			wantCol: "created_at",
			wantDir: "asc",
		},
		{
			name:    "invalid dir falls back to asc",
			by:      "name",
			dir:     "sideways",
			wantCol: "c.name",
			wantDir: "asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, dir := ResolveSort(allow, tt.by, tt.dir)
			if col != tt.wantCol {
				t.Errorf("col = %q, want %q", col, tt.wantCol)
			}
			if dir != tt.wantDir {
				t.Errorf("dir = %q, want %q", dir, tt.wantDir)
			}
		})
	}
}
