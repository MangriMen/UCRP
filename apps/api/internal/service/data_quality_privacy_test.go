package service

import "testing"

func TestVisibleMergeCount(t *testing.T) {
	for _, tc := range []struct {
		pending, guests int
		visible         bool
	}{{4, 5, false}, {5, 4, false}, {5, 5, true}} {
		value := visibleMergeCount(tc.pending, tc.guests)
		if (value != nil) != tc.visible || value != nil && *value != tc.pending {
			t.Fatalf("pending=%d guests=%d: got %v", tc.pending, tc.guests, value)
		}
	}
}
