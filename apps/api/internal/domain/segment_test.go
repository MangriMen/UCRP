package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSegmentRule_UnmarshalKeepsPurchasedZones(t *testing.T) {
	var rule SegmentRule
	payload := `{"purchasedZones":[{"mallId":"m1","floor":2},{"mallId":"m2","floor":-1}],"periodDays":30}`
	if err := json.Unmarshal([]byte(payload), &rule); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := []ZoneRef{{MallID: "m1", Floor: 2}, {MallID: "m2", Floor: -1}}
	if !reflect.DeepEqual(rule.PurchasedZones, want) {
		t.Fatalf("PurchasedZones = %v, want %v", rule.PurchasedZones, want)
	}
}

func TestSegmentRule_GroupRoundTrip(t *testing.T) {
	payload := `{"operator":"and","conditions":[{"visitedTenantIds":["t1"],"periodDays":30},{"operator":"or","conditions":[{"purchasedZones":[{"mallId":"m1","floor":2}],"periodDays":7}]}]}`
	var rule SegmentRule
	if err := json.Unmarshal([]byte(payload), &rule); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !rule.IsGroup() || len(rule.Conditions) != 2 || !rule.Conditions[1].IsGroup() {
		t.Fatalf("group structure lost: %+v", rule)
	}
	encoded, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded SegmentRule
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal round trip: %v", err)
	}
	if !reflect.DeepEqual(decoded, rule) {
		t.Fatalf("round trip = %+v, want %+v", decoded, rule)
	}
}

func TestSegmentRule_RoundTripPreservesPurchasedZones(t *testing.T) {
	original := SegmentRule{PurchasedZones: []ZoneRef{{MallID: "m1", Floor: 3}}, PeriodDays: 14}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SegmentRule
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("round trip = %+v, want %+v", decoded, original)
	}
}
