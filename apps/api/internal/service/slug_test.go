package service

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Island & Coastal Tours":            "island-coastal-tours",
		"Du lịch Đà Nẵng – Hội An":          "du-lich-da-nang-hoi-an",
		"  Núi  &  Trekking!!  ":            "nui-trekking",
		"Đảo Phú Quốc":                      "dao-phu-quoc",
		"Café  Crème":                       "cafe-creme",
		"already-a-slug":                    "already-a-slug",
		"UPPER Case 123":                    "upper-case-123",
		"---leading and trailing---":        "leading-and-trailing",
		"":                                  "",
		"!!!":                               "",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSlugifyCapsLengthWithoutTrailingDash(t *testing.T) {
	long := strings.Repeat("ab ", 100)
	got := Slugify(long)
	if len(got) > maxSlugLen {
		t.Fatalf("len = %d, want <= %d", len(got), maxSlugLen)
	}
	if strings.HasSuffix(got, "-") {
		t.Fatalf("capped slug must not end in a dash: %q", got)
	}
}
