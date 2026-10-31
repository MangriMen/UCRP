package domain

import "testing"

func TestMaskIdentifierValue(t *testing.T) {
	cases := map[string]struct {
		idType string
		value  string
		want   string
	}{
		"phone keeps the last four digits":            {"phone", "79001234567", "***4567"},
		"card keeps the last four digits":             {"card", "4276123412341234", "***1234"},
		"email keeps the first letter and the domain": {"email", "ivan.petrov@example.com", "i***@example.com"},
		"email without a local part keeps the tail":   {"email", "@example.com", "***.com"},
		"short value is masked whole":                 {"loyalty_id", "1234567", "***"},
		"non latin value keeps whole runes":           {"car_plate", "А123ВС777", "***С777"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := MaskIdentifierValue(tc.idType, tc.value); got != tc.want {
				t.Fatalf("MaskIdentifierValue(%q, %q) = %q, want %q", tc.idType, tc.value, got, tc.want)
			}
		})
	}
}
