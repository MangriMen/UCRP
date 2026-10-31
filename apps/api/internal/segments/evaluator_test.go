package segments

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestValidateRule(t *testing.T) {
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"t1"}}, 0); err != nil {
		t.Fatalf("valid rule rejected: %v", err)
	}
	if err := validateRule(domain.SegmentRule{}, 0); err == nil {
		t.Fatal("empty was accepted")
	}
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"a"}, ExcludedTenantIDs: []string{"a"}}, 0); err == nil {
		t.Fatal("overlap was/excluded_tenant_ids accepted")
	}
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"a"}, RequireContact: "phone"}, 0); err != nil {
		t.Fatalf("require_contact phone rejected: %v", err)
	}
	if err := validateRule(domain.SegmentRule{VisitedTenantIDs: []string{"a"}, RequireContact: "carrier_pigeon"}, 0); err == nil {
		t.Fatal("bad require_contact accepted")
	}
}

func TestValidateRule_AcceptsZonesWithoutTenants(t *testing.T) {
	rule := domain.SegmentRule{PurchasedZones: []domain.ZoneRef{{MallID: "m1", Floor: 2}, {MallID: "m2", Floor: 2}}}
	if err := validateRule(rule, 0); err != nil {
		t.Fatalf("zones-only rule rejected: %v", err)
	}
}

func TestValidateRule_RejectsInvalidZones(t *testing.T) {
	cases := map[string][]domain.ZoneRef{
		"empty mall":     {{MallID: "", Floor: 1}},
		"duplicate zone": {{MallID: "m1", Floor: 1}, {MallID: "m1", Floor: 1}},
	}
	for name, zones := range cases {
		if err := validateRule(domain.SegmentRule{PurchasedZones: zones}, 0); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestValidateRule_AcceptsGroundFloorZone(t *testing.T) {
	rule := domain.SegmentRule{PurchasedZones: []domain.ZoneRef{{MallID: "m1", Floor: 0}}}
	if err := validateRule(rule, 0); err != nil {
		t.Fatalf("floor 0 zone rejected: %v", err)
	}
}

func TestBuildQuery_SkipsInactiveMalls(t *testing.T) {
	rule := domain.SegmentRule{PurchasedZones: []domain.ZoneRef{{MallID: "m1", Floor: 1}}}
	query, _ := buildQuery(rule, time.Unix(0, 0))
	if !strings.Contains(query, "mz.is_active") {
		t.Fatalf("zone query must exclude inactive malls: %s", query)
	}
}

func TestBuildQuery_UsesPurchasesOnlyForZones(t *testing.T) {
	rule := domain.SegmentRule{PurchasedZones: []domain.ZoneRef{{MallID: "m1", Floor: 2}, {MallID: "m1", Floor: 3}}}
	query, args := buildQuery(rule, time.Unix(0, 0))

	if !strings.Contains(query, "ept.code = 'purchase'") {
		t.Fatalf("zones must be matched by purchases: %s", query)
	}
	if strings.Contains(query, "counts_as_presence") {
		t.Fatalf("zones-only rule must not use presence events: %s", query)
	}
	wantArgs := []any{time.Unix(0, 0), "m1", 2, "m1", 3}
	if len(args) != len(wantArgs) {
		t.Fatalf("args = %v, want %v", args, wantArgs)
	}
	for i := range wantArgs {
		if args[i] != wantArgs[i] {
			t.Fatalf("args[%d] = %v, want %v", i, args[i], wantArgs[i])
		}
	}
}

func TestBuildQuery_CombinesTenantsZonesExclusionsAndContact(t *testing.T) {
	rule := domain.SegmentRule{
		VisitedTenantIDs:  []string{"t1", "t2"},
		ExcludedTenantIDs: []string{"t3"},
		PurchasedZones:    []domain.ZoneRef{{MallID: "m1", Floor: 1}},
		RequireContact:    "phone",
	}
	query, args := buildQuery(rule, time.Unix(0, 0))

	for _, fragment := range []string{
		"e.tenant_id IN ($2,$3)",
		"ep.guest_id = base.guest_id",
		"(tz.mall_id = $4 AND tz.floor = $5)",
		"e2.tenant_id IN ($6)",
		"i.id_type = $7",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("query misses %q:\n%s", fragment, query)
		}
	}
	if len(args) != 7 {
		t.Fatalf("len(args) = %d, want 7", len(args))
	}
}

func TestBuildQuery_KeepsTenantRuleWithoutZones(t *testing.T) {
	query, args := buildQuery(domain.SegmentRule{VisitedTenantIDs: []string{"t1"}}, time.Unix(0, 0))

	if strings.Contains(query, "tenant tz") {
		t.Fatalf("tenant-only rule must not join zones: %s", query)
	}
	if strings.Contains(query, "WHERE base") || strings.Contains(query, ") base WHERE") {
		t.Fatalf("tenant-only rule must not filter the base set: %s", query)
	}
	if len(args) != 2 {
		t.Fatalf("len(args) = %d, want 2", len(args))
	}
}

func TestValidateRuleGroups(t *testing.T) {
	leaf := domain.SegmentRule{VisitedTenantIDs: []string{"a"}}
	deep := leaf
	for range maxRuleDepth + 1 {
		deep = domain.SegmentRule{Operator: "and", Conditions: []domain.SegmentRule{deep}}
	}
	cases := map[string]domain.SegmentRule{
		"empty conditions": {Operator: "and"},
		"unknown operator": {Operator: "xor", Conditions: []domain.SegmentRule{leaf}},
		"invalid leaf":     {Operator: "or", Conditions: []domain.SegmentRule{leaf, {}}},
		"too deep":         deep,
	}
	for name, rule := range cases {
		if err := validateRule(rule, 0); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
	nested := domain.SegmentRule{Operator: "and", Conditions: []domain.SegmentRule{
		leaf,
		{Operator: "or", Conditions: []domain.SegmentRule{leaf, {PurchasedZones: []domain.ZoneRef{{MallID: "m", Floor: 1}}}}},
	}}
	if err := validateRule(nested, 0); err != nil {
		t.Fatalf("nested group: %v", err)
	}
}

func TestCombineGuestIDs(t *testing.T) {
	left := []string{"a", "b", "c"}
	right := []string{"b", "c", "d"}
	if got := combineGuestIDs("and", left, right); !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Errorf("and: %v", got)
	}
	got := combineGuestIDs("or", left, right)
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Errorf("or: %v", got)
	}
}
