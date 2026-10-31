package domain

import (
	"reflect"
	"testing"
)

func TestMall_Floors(t *testing.T) {
	cases := map[string]struct {
		mall Mall
		want []int
	}{
		"skips zero by default":   {Mall{FloorMin: -1, FloorMax: 3}, []int{-1, 1, 2, 3}},
		"keeps zero when enabled": {Mall{FloorMin: -1, FloorMax: 2, HasZero: true}, []int{-1, 0, 1, 2}},
		"single floor":            {Mall{FloorMin: 1, FloorMax: 1}, []int{1}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.mall.Floors(); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestMall_ValidFloorRange(t *testing.T) {
	cases := map[string]struct {
		mall Mall
		want bool
	}{
		"ordered":  {Mall{FloorMin: -1, FloorMax: 3}, true},
		"inverted": {Mall{FloorMin: 3, FloorMax: 1}, false},
		"too wide": {Mall{FloorMin: -500, FloorMax: 500}, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.mall.ValidFloorRange(); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestMall_HasFloor(t *testing.T) {
	cases := map[string]struct {
		mall  Mall
		floor int
		want  bool
	}{
		"below range":        {Mall{FloorMin: -1, FloorMax: 3}, -2, false},
		"lower bound":        {Mall{FloorMin: -1, FloorMax: 3}, -1, true},
		"upper bound":        {Mall{FloorMin: -1, FloorMax: 3}, 3, true},
		"above range":        {Mall{FloorMin: -1, FloorMax: 3}, 4, false},
		"zero when skipped":  {Mall{FloorMin: -1, FloorMax: 3}, 0, false},
		"zero when enabled":  {Mall{FloorMin: -1, FloorMax: 3, HasZero: true}, 0, true},
		"zero outside range": {Mall{FloorMin: 1, FloorMax: 3, HasZero: true}, 0, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.mall.HasFloor(tc.floor); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
