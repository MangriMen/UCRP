package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type fakeLimits struct {
	items []domain.ContactLimit
}

func (f *fakeLimits) List(context.Context, int, int) ([]domain.ContactLimit, error) {
	return f.items, nil
}
func (f *fakeLimits) ListActive(context.Context) ([]domain.ContactLimit, error) {
	return f.items, nil
}
func (f *fakeLimits) GetByID(context.Context, string) (*domain.ContactLimit, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeLimits) Insert(context.Context, *domain.ContactLimit) error { return nil }
func (f *fakeLimits) Update(context.Context, *domain.ContactLimit) error { return nil }
func (f *fakeLimits) Delete(context.Context, string) error               { return nil }

type fakeContacts struct {
	all    map[string]int
	byChan map[string]map[string]int
}

func (f *fakeContacts) ListByGuest(context.Context, string, int, int) ([]domain.ContactLog, error) {
	return nil, nil
}
func (f *fakeContacts) CountSince(_ context.Context, ids []string, _ time.Time) (map[string]int, error) {
	out := map[string]int{}
	for _, id := range ids {
		out[id] = f.all[id]
	}
	return out, nil
}
func (f *fakeContacts) CountSinceChannel(_ context.Context, ids []string, _ time.Time, channel string) (map[string]int, error) {
	out := map[string]int{}
	m := f.byChan[channel]
	for _, id := range ids {
		out[id] = m[id]
	}
	return out, nil
}
func (f *fakeContacts) InsertBatch(context.Context, []domain.ContactLog) error { return nil }

func TestBlockedForChannelAnyRuleBlocks(t *testing.T) {
	svc := &ExportService{
		limits: &fakeLimits{items: []domain.ContactLimit{
			{ID: "r1", Channel: "sms", MaxContacts: 2, PeriodDays: 7, IsActive: true},
			{ID: "r2", Channel: "sms", MaxContacts: 1, PeriodDays: 30, IsActive: true},
		}},
		contacts: &fakeContacts{
			all:    map[string]int{},
			byChan: map[string]map[string]int{"sms": {"g1": 2, "g2": 0}},
		},
		clock: domain.SystemClock{},
	}
	blocked, err := svc.blockedForChannel(context.Background(), []memberCandidate{{guestID: "g1"}, {guestID: "g2"}}, "sms")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocked) != 1 || blocked[0] != "g1" {
		t.Fatalf("blocked=%v want [g1]", blocked)
	}
}

func TestBlockedForChannelGlobalCountsAllChannels(t *testing.T) {
	svc := &ExportService{
		limits: &fakeLimits{items: []domain.ContactLimit{
			{ID: "g1", Channel: "", MaxContacts: 2, PeriodDays: 7, IsActive: true},
		}},
		contacts: &fakeContacts{
			all:    map[string]int{"guest-x": 2},
			byChan: map[string]map[string]int{"sms": {"guest-x": 1}},
		},
		clock: domain.SystemClock{},
	}
	blocked, err := svc.blockedForChannel(context.Background(), []memberCandidate{{guestID: "guest-x"}}, "sms")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocked) != 1 {
		t.Fatalf("global limit should block with cross-channel count, got %v", blocked)
	}
}

func TestBlockedForChannelStrictestRuleWins(t *testing.T) {
	svc := &ExportService{
		limits: &fakeLimits{items: []domain.ContactLimit{
			{ID: "r1", Channel: "sms", MaxContacts: 10, PeriodDays: 7, IsActive: true},
			{ID: "r2", Channel: "sms", MaxContacts: 1, PeriodDays: 30, IsActive: true},
		}},
		contacts: &fakeContacts{
			all: map[string]int{},
			byChan: map[string]map[string]int{
				"sms": {"guest-y": 2},
			},
		},
		clock: domain.SystemClock{},
	}
	blocked, err := svc.blockedForChannel(context.Background(), []memberCandidate{{guestID: "guest-y"}}, "sms")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocked) != 1 {
		t.Fatalf("strictest rule should block, got %v", blocked)
	}
}

func TestSegmentNotReadyConflict(t *testing.T) {
	svc := &ExportService{
		exports:  &fakeExportRepo{},
		limits:   &fakeLimits{},
		contacts: &fakeContacts{},
		segments: &fakeSegments{status: domain.VersionComputing},
		clock:    domain.SystemClock{},
	}
	_, err := svc.Create(context.Background(), ExportCreateRequest{
		SegmentID: "seg", Channel: "sms", Format: "csv", IdempotencyKey: "k1",
	}, "u1")
	var ce *ExportConflictError
	if err == nil {
		t.Fatal("expected conflict")
	}
	if !errors.As(err, &ce) || ce.Code != "segment_not_ready" {
		t.Fatalf("err=%v", err)
	}
}

func TestAllBlockedConflict(t *testing.T) {
	svc := &ExportService{
		exports:  &fakeExportRepo{},
		limits:   &fakeLimits{items: []domain.ContactLimit{{Channel: "sms", MaxContacts: 0, PeriodDays: 1, IsActive: true}}},
		contacts: &fakeContacts{all: map[string]int{}, byChan: map[string]map[string]int{"sms": {"g1": 5}}},
		segments: &fakeSegments{status: domain.VersionReady, memberContacts: []domain.GuestContact{{GuestID: "g1", Contact: "7900", IDType: "phone"}}},
		clock:    domain.SystemClock{},
	}
	// MaxContacts 0 invalid-ish; use max 1 with count 2 instead
	svc.limits = &fakeLimits{items: []domain.ContactLimit{{Channel: "sms", MaxContacts: 1, PeriodDays: 7, IsActive: true}}}
	svc.contacts = &fakeContacts{all: map[string]int{}, byChan: map[string]map[string]int{"sms": {"g1": 5}}}
	_, err := svc.Create(context.Background(), ExportCreateRequest{
		SegmentID: "seg", Channel: "sms", Format: "csv", IdempotencyKey: "k2",
	}, "u1")
	var ce *ExportConflictError
	if !errors.As(err, &ce) || ce.Code != "contact_limit_exceeded" {
		t.Fatalf("want contact_limit_exceeded, got %v", err)
	}
}

type fakeExportRepo struct{}

func (f *fakeExportRepo) FindByKey(context.Context, string, string) (*domain.ExportJob, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeExportRepo) GetByID(context.Context, string) (*domain.ExportJob, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeExportRepo) ListMembers(context.Context, string) ([]domain.ExportMember, error) {
	return nil, nil
}
func (f *fakeExportRepo) Insert(context.Context, *domain.ExportJob, []domain.ExportMember) error {
	return nil
}
func (f *fakeExportRepo) ListMemberPage(context.Context, string, *bool, int, int) ([]domain.ExportMember, int, error) {
	return nil, 0, nil
}
func (f *fakeExportRepo) List(context.Context, domain.ExportFilter, int, int) ([]domain.ExportJob, int, error) {
	return nil, 0, nil
}

type fakeSegments struct {
	status         domain.SegmentVersionStatus
	memberContacts []domain.GuestContact
}

func (f *fakeSegments) Insert(context.Context, *domain.Segment, *domain.SegmentVersion) error {
	return nil
}
func (f *fakeSegments) GetByID(context.Context, string) (*domain.Segment, *domain.SegmentVersion, error) {
	return &domain.Segment{ID: "seg"}, &domain.SegmentVersion{ID: "v1", Status: f.status}, nil
}
func (f *fakeSegments) List(context.Context, domain.SegmentFilter, int, int) ([]domain.Segment, int, error) {
	return nil, 0, nil
}
func (f *fakeSegments) ListTemplates(context.Context) ([]domain.Segment, error) {
	return nil, nil
}
func (f *fakeSegments) ListVersions(context.Context, string) ([]domain.SegmentVersion, error) {
	return nil, nil
}
func (f *fakeSegments) GetVersion(context.Context, string, int) (*domain.Segment, *domain.SegmentVersion, error) {
	return nil, nil, domain.ErrNotFound
}
func (f *fakeSegments) InsertVersion(context.Context, *domain.SegmentVersion) error { return nil }
func (f *fakeSegments) Archive(context.Context, string, time.Time) error            { return nil }
func (f *fakeSegments) MarkComputing(context.Context, string) error                 { return nil }
func (f *fakeSegments) ReplaceMembers(context.Context, string, []string, time.Time) error {
	return nil
}
func (f *fakeSegments) ListMembers(context.Context, string, int, int) ([]domain.SegmentMemberEntry, error) {
	return nil, nil
}
func (f *fakeSegments) CountMembers(context.Context, string) (int, error) {
	return 0, nil
}
func (f *fakeSegments) CurrentVersions(context.Context, []string) (map[string]*domain.SegmentVersion, error) {
	return map[string]*domain.SegmentVersion{}, nil
}
func (f *fakeSegments) ListMemberContacts(context.Context, string) ([]domain.GuestContact, error) {
	return f.memberContacts, nil
}
