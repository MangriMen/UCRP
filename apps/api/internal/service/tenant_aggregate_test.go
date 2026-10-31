package service

import "testing"

func TestRepeatVisitGuests_HidesRepeats_WhenFirstVisitsAreBelowThreshold(t *testing.T) {
	reveal := &aggregateReveal{threshold: 5}
	visits, repeats, repeatGuests := 43, 40, 30
	if got := reveal.count(repeats, repeatVisitGuests(repeatGuests, visits-repeats)); got != nil {
		t.Fatalf("repeat visits = %d, want hidden: visits minus repeats would reveal three new guests", *got)
	}
	if !reveal.suppressed {
		t.Fatal("hiding repeat visits must mark the aggregate as suppressed")
	}
}

func TestRepeatVisitGuests_RevealsRepeats_WhenEveryVisitIsARepeat(t *testing.T) {
	reveal := &aggregateReveal{threshold: 5}
	if got := reveal.count(12, repeatVisitGuests(8, 0)); got == nil || *got != 12 {
		t.Fatalf("repeat visits = %v, want 12", got)
	}
}

func TestRepeatVisitGuests_RevealsRepeats_WhenFirstVisitsReachThreshold(t *testing.T) {
	reveal := &aggregateReveal{threshold: 5}
	if got := reveal.count(5, repeatVisitGuests(5, 5)); got == nil || *got != 5 {
		t.Fatalf("repeat visits = %v, want 5", got)
	}
}
