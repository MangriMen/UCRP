package segments

import (
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestValidateRule(t *testing.T) {
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"t1"}}); err != nil {
		t.Fatalf("valid rule rejected: %v", err)
	}
	if err := validateRule(domain.SegmentRule{}); err == nil {
		t.Fatal("empty was accepted")
	}
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"a"}, ExcludedTenantIDs: []string{"a"}}); err == nil {
		t.Fatal("overlap was/excluded_tenant_ids accepted")
	}
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"a"}, RequireContact: "phone"}); err != nil {
		t.Fatalf("require_contact phone rejected: %v", err)
	}
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"a"}, RequireContact: "carrier_pigeon"}); err == nil {
		t.Fatal("bad require_contact accepted")
	}
}
