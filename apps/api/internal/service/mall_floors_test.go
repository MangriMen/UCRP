package service

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

func intPtr(v int) *int { return &v }

func TestMallRequestApplyFloors(t *testing.T) {
	tests := []struct {
		name    string
		req     MallRequest
		wantErr bool
	}{
		{"valid range", MallRequest{FloorMin: intPtr(-1), FloorMax: intPtr(3)}, false},
		{"inverted range", MallRequest{FloorMin: intPtr(3), FloorMax: intPtr(1)}, true},
		{"too wide", MallRequest{FloorMin: intPtr(-200), FloorMax: intPtr(200)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &domain.Mall{FloorMin: 1, FloorMax: 1}
			err := tt.req.applyFloors(m)
			if tt.wantErr != errors.Is(err, domain.ErrValidation) {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTenantUpdateRequestFloor(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantSet   bool
		wantFloor *int
	}{
		{"absent", `{}`, false, nil},
		{"null clears", `{"floor":null}`, true, nil},
		{"zero is a floor", `{"floor":0}`, true, intPtr(0)},
		{"value", `{"floor":2}`, true, intPtr(2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req TenantUpdateRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatal(err)
			}
			if req.Floor.Set != tt.wantSet {
				t.Fatalf("Set = %v, want %v", req.Floor.Set, tt.wantSet)
			}
			got, want := req.Floor.Value, tt.wantFloor
			if (got == nil) != (want == nil) || (got != nil && *got != *want) {
				t.Fatalf("Value = %v, want %v", got, want)
			}
		})
	}
}
