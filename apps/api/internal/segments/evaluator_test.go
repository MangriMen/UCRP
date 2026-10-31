package segments

import (
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestValidateRule(t *testing.T) {
	if err := validateRule(domain.SegmentRule{Was: []string{"t1"}}); err != nil {
		t.Fatalf("valid rule rejected: %v", err)
	}
	if err := validateRule(domain.SegmentRule{}); err == nil {
		t.Fatal("empty was accepted")
	}
	if err := validateRule(domain.SegmentRule{Was: []string{"a"}, WasNot: []string{"a"}}); err == nil {
		t.Fatal("overlap was/was_not accepted")
	}
	if err := validateRule(domain.SegmentRule{Was: []string{"a"}, RequireContact: "phone"}); err != nil {
		t.Fatalf("require_contact phone rejected: %v", err)
	}
	if err := validateRule(domain.SegmentRule{Was: []string{"a"}, RequireContact: "carrier_pigeon"}); err == nil {
		t.Fatal("bad require_contact accepted")
	}
}
