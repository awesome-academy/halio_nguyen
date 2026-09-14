package service

import "testing"

func TestEffectivePrice(t *testing.T) {
	price := 100.0
	discount := 80.0
	override := 60.0

	cases := []struct {
		name     string
		override *float64
		discount *float64
		price    float64
		want     float64
	}{
		{"override wins over discount and price", &override, &discount, price, 60.0},
		{"discount wins when no override", nil, &discount, price, 80.0},
		{"base price when neither set", nil, nil, price, 100.0},
		{"override wins even with no discount", &override, nil, price, 60.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EffectivePrice(tc.override, tc.discount, tc.price)
			if got != tc.want {
				t.Errorf("EffectivePrice() = %v, want %v", got, tc.want)
			}
		})
	}
}
