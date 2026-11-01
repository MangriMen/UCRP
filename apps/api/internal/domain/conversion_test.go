package domain

import (
	"math"
	"testing"
)

func TestConversion_ComputesLiftAndPValue_WhenBothGroupsPresent(t *testing.T) {
	c := Conversion{Recipients: 1000, Conversions: 120, ControlGroup: 1000, ControlConversions: 80}
	if c.Rate() != 0.12 || c.ControlRate() != 0.08 {
		t.Fatalf("rates: %v %v", c.Rate(), c.ControlRate())
	}
	if math.Abs(c.Lift()-0.04) > 1e-9 {
		t.Fatalf("lift: %v", c.Lift())
	}
	p, ok := c.PValue()
	if !ok || p <= 0.002 || p >= 0.004 {
		t.Fatalf("p-value: %v %v", p, ok)
	}
}

func TestConversion_HasNoPValueOrLift_WhenControlGroupEmpty(t *testing.T) {
	c := Conversion{Recipients: 10, Conversions: 3}
	if _, ok := c.PValue(); ok {
		t.Fatal("p-value without control group")
	}
	if c.Lift() != 0 || c.Rate() != 0.3 {
		t.Fatalf("lift %v rate %v", c.Lift(), c.Rate())
	}
}
