package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type ExportService struct {
	exports  domain.ExportRepo
	limits   domain.ContactLimitRepo
	contacts domain.ContactLogRepo
	segments domain.SegmentRepo
	idGen    domain.IDGen
	clock    domain.Clock
}

func NewExportService(
	exports domain.ExportRepo,
	limits domain.ContactLimitRepo,
	contacts domain.ContactLogRepo,
	segments domain.SegmentRepo,
) *ExportService {
	return &ExportService{
		exports:  exports,
		limits:   limits,
		contacts: contacts,
		segments: segments,
		idGen:    idgen.UUID{},
		clock:    domain.SystemClock{},
	}
}

type ExportPreviewRequest struct {
	SegmentID      string `json:"segment_id"`
	Channel        string `json:"channel"`
	ControlPercent int    `json:"control_percent"`
	IdempotencyKey string `json:"idempotency_key"`
}

type ExportPreviewResponse struct {
	InSegment      int `json:"in_segment"`
	WithContact    int `json:"with_contact"`
	BlockedByLimit int `json:"blocked_by_limit"`
	Control        int `json:"control"`
	ToSend         int `json:"to_send"`
}

type ExportCreateRequest struct {
	SegmentID      string `json:"segment_id"`
	Channel        string `json:"channel"`
	Format         string `json:"format"`
	ControlPercent int    `json:"control_percent"`
	CampaignID     string `json:"campaign_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

type ExportDTO struct {
	ID                  string `json:"id"`
	SegmentID           string `json:"segment_id"`
	Channel             string `json:"channel"`
	Format              string `json:"format"`
	Status              string `json:"status"`
	TotalInSegment      int    `json:"total_in_segment"`
	NoContactCount      int    `json:"no_contact_count"`
	BlockedByLimitCount int    `json:"blocked_by_limit_count"`
	ControlCount        int    `json:"control_count"`
	ExportedCount       int    `json:"exported_count"`
}

type memberCandidate struct {
	guestID      string
	identifierID string
	contact      string
	idType       string
}

func (s *ExportService) Preview(ctx context.Context, req ExportPreviewRequest) (*ExportPreviewResponse, error) {
	if err := validateChannel(req.Channel); err != nil {
		return nil, err
	}
	if req.SegmentID == "" || req.ControlPercent < 0 || req.ControlPercent > 50 {
		return nil, domain.ErrValidation
	}
	all, withContact, err := s.memberLists(ctx, req.SegmentID, req.Channel)
	if err != nil {
		return nil, err
	}
	withContact = dedupeByGuest(withContact)
	blocked, err := s.blockedForChannel(ctx, withContact, req.Channel)
	if err != nil {
		return nil, err
	}
	allowed := withoutIDs(withContact, blocked)
	control := pickControl(allowed, req.ControlPercent, hashSeed(req.IdempotencyKey))
	totalGuests := len(dedupeByGuest(all))
	return &ExportPreviewResponse{
		InSegment:      totalGuests,
		WithContact:    len(withContact),
		BlockedByLimit: len(blocked),
		Control:        len(control),
		ToSend:         len(allowed) - len(control),
	}, nil
}

func (s *ExportService) Create(ctx context.Context, req ExportCreateRequest, actorUserID string) (*ExportDTO, error) {
	if err := validateExportRequest(req.SegmentID, req.Channel, req.Format, req.ControlPercent, req.IdempotencyKey); err != nil {
		return nil, err
	}
	if existing, err := s.exports.FindByKey(ctx, actorUserID, req.IdempotencyKey); err == nil {
		return toExportDTO(*existing), nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	all, withContact, err := s.memberLists(ctx, req.SegmentID, req.Channel)
	if err != nil {
		return nil, err
	}
	_, ver, err := s.segments.GetByID(ctx, req.SegmentID)
	if err != nil {
		return nil, err
	}
	withContact = dedupeByGuest(withContact)
	blocked, err := s.blockedForChannel(ctx, withContact, req.Channel)
	if err != nil {
		return nil, err
	}
	allowed := withoutIDs(withContact, blocked)
	controlIDs := pickControl(allowed, req.ControlPercent, hashSeed(req.IdempotencyKey))
	controlSet := map[string]bool{}
	for _, c := range controlIDs {
		controlSet[c.guestID] = true
	}

	now := s.clock.Now()
	totalGuests := len(dedupeByGuest(all))
	job := domain.ExportJob{
		ID:                  s.idGen.NewID(ctx),
		SegmentID:           req.SegmentID,
		SegmentVersionID:    ver.ID,
		CampaignID:          req.CampaignID,
		Channel:             domain.ExportChannel(req.Channel),
		Format:              domain.ExportFormat(orDefault(req.Format, string(domain.FormatCSV))),
		ControlPercent:      req.ControlPercent,
		ControlSeed:         hashSeed(req.IdempotencyKey),
		IdempotencyKey:      req.IdempotencyKey,
		Status:              domain.ExportReady,
		TotalInSegment:      totalGuests,
		NoContactCount:      totalGuests - len(withContact),
		BlockedByLimitCount: len(blocked),
		ControlCount:        len(controlSet),
		ExportedCount:       len(allowed) - len(controlSet),
		CreatedBy:           actorUserID,
		CreatedAt:           now,
		FinishedAt:          &now,
	}

	exportMembers := make([]domain.ExportMember, 0, len(allowed))
	logs := make([]domain.ContactLog, 0, job.ExportedCount)
	for _, c := range allowed {
		isControl := controlSet[c.guestID]
		exportMembers = append(exportMembers, domain.ExportMember{
			ExportID:     job.ID,
			GuestID:      c.guestID,
			IdentifierID: c.identifierID,
			IsControl:    isControl,
			Contact:      c.contact,
		})
		if !isControl {
			logs = append(logs, domain.ContactLog{
				ID:          s.idGen.NewID(ctx),
				GuestID:     c.guestID,
				ExportID:    job.ID,
				ContactedAt: now,
			})
		}
	}
	if err := s.exports.Insert(ctx, &job, exportMembers); err != nil {
		return nil, err
	}
	if err := s.contacts.InsertBatch(ctx, logs); err != nil {
		return nil, err
	}
	return toExportDTO(job), nil
}

func (s *ExportService) CSV(ctx context.Context, exportID string) ([]byte, error) {
	job, err := s.exports.GetByID(ctx, exportID)
	if err != nil {
		return nil, err
	}
	if job.Status != domain.ExportReady {
		return nil, domain.ErrConflict
	}
	members, err := s.exports.ListMembers(ctx, exportID)
	if err != nil {
		return nil, err
	}
	buf := &strings.Builder{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"guest_id", "contact"})
	for _, m := range members {
		if m.IsControl {
			continue
		}
		_ = w.Write([]string{m.GuestID, m.Contact})
	}
	w.Flush()
	return []byte(buf.String()), w.Error()
}

func (s *ExportService) List(ctx context.Context, limit, offset int) ([]ExportDTO, error) {
	items, err := s.exports.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]ExportDTO, 0, len(items))
	for _, j := range items {
		out = append(out, *toExportDTO(j))
	}
	return out, nil
}

func (s *ExportService) Get(ctx context.Context, id string) (*ExportDTO, error) {
	job, err := s.exports.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toExportDTO(*job), nil
}

func (s *ExportService) memberLists(ctx context.Context, segmentID, channel string) (all, withContact []memberCandidate, err error) {
	_, ver, err := s.segments.GetByID(ctx, segmentID)
	if err != nil {
		return nil, nil, err
	}
	contacts, err := s.segments.ListMemberContacts(ctx, ver.ID)
	if err != nil {
		return nil, nil, err
	}
	all = toCandidates(contacts)
	withContact = filterChannel(all, channel)
	return all, withContact, nil
}

func (s *ExportService) blockedForChannel(ctx context.Context, candidates []memberCandidate, channel string) ([]string, error) {
	limits, err := s.limits.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	applicable := make([]domain.ContactLimit, 0, len(limits))
	for _, l := range limits {
		if l.Channel == "" || l.Channel == channel {
			applicable = append(applicable, l)
		}
	}
	limits = applicable
	if len(limits) == 0 || len(candidates) == 0 {
		return nil, nil
	}
	ids := make([]string, len(candidates))
	for i, c := range candidates {
		ids[i] = c.guestID
	}
	period := limits[0].PeriodDays
	maxN := limits[0].MaxContacts
	for _, l := range limits {
		if l.PeriodDays < period {
			period = l.PeriodDays
		}
		if l.MaxContacts < maxN {
			maxN = l.MaxContacts
		}
	}
	since := s.clock.Now().UTC().AddDate(0, 0, -period)
	counts, err := s.contacts.CountSinceChannel(ctx, ids, since, channel)
	if err != nil {
		return nil, err
	}
	var blocked []string
	for _, c := range candidates {
		if counts[c.guestID] >= maxN {
			blocked = append(blocked, c.guestID)
		}
	}
	return blocked, nil
}

func toCandidates(in []domain.GuestContact) []memberCandidate {
	out := make([]memberCandidate, 0, len(in))
	for _, c := range in {
		out = append(out, memberCandidate{
			guestID:      c.GuestID,
			identifierID: c.IdentifierID,
			contact:      c.Contact,
			idType:       c.IDType,
		})
	}
	return out
}

func filterChannel(in []memberCandidate, channel string) []memberCandidate {
	needPhone := channel == string(domain.ChannelSMS) || channel == string(domain.ChannelPush)
	out := make([]memberCandidate, 0, len(in))
	for _, c := range in {
		if c.contact == "" {
			continue
		}
		if needPhone && c.idType == "phone" {
			out = append(out, c)
		}
		if !needPhone && c.idType == "email" {
			out = append(out, c)
		}
	}
	return out
}

func withoutIDs(in []memberCandidate, blocked []string) []memberCandidate {
	set := map[string]bool{}
	for _, id := range blocked {
		set[id] = true
	}
	out := make([]memberCandidate, 0, len(in))
	for _, c := range in {
		if !set[c.guestID] {
			out = append(out, c)
		}
	}
	return out
}

func pickControl(in []memberCandidate, percent int, seed int64) []memberCandidate {
	if percent <= 0 || len(in) == 0 {
		return nil
	}
	n := (len(in) * percent) / 100
	if n < 1 {
		n = 1
	}
	if n > len(in) {
		n = len(in)
	}
	indexes := make([]int, len(in))
	for i := range indexes {
		indexes[i] = i
	}
	for i := len(indexes) - 1; i > 0; i-- {
		j := int(hash64(in[i].guestID, seed+int64(i)) % uint64(i+1))
		indexes[i], indexes[j] = indexes[j], indexes[i]
	}
	out := make([]memberCandidate, 0, n)
	for _, idx := range indexes[:n] {
		out = append(out, in[idx])
	}
	return out
}

func hashSeed(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int64(h.Sum64())
}

func hash64(guestID string, seed int64) uint64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s:%d", guestID, seed)
	return h.Sum64()
}

func toExportDTO(j domain.ExportJob) *ExportDTO {
	return &ExportDTO{
		ID:                  j.ID,
		SegmentID:           j.SegmentID,
		Channel:             string(j.Channel),
		Format:              string(j.Format),
		Status:              string(j.Status),
		TotalInSegment:      j.TotalInSegment,
		NoContactCount:      j.NoContactCount,
		BlockedByLimitCount: j.BlockedByLimitCount,
		ControlCount:        j.ControlCount,
		ExportedCount:       j.ExportedCount,
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func validateChannel(channel string) error {
	switch channel {
	case string(domain.ChannelSMS), string(domain.ChannelPush), string(domain.ChannelEmail):
		return nil
	default:
		return domain.ErrValidation
	}
}

func validateExportRequest(segmentID, channel, format string, percent int, idemKey string) error {
	if segmentID == "" {
		return domain.ErrValidation
	}
	if err := validateChannel(channel); err != nil {
		return err
	}
	if format != "" && format != string(domain.FormatCSV) && format != string(domain.FormatAPI) {
		return domain.ErrValidation
	}
	if percent < 0 || percent > 50 {
		return domain.ErrValidation
	}
	if idemKey == "" || len(idemKey) > 64 {
		return domain.ErrValidation
	}
	return nil
}

func dedupeByGuest(in []memberCandidate) []memberCandidate {
	seen := map[string]bool{}
	out := make([]memberCandidate, 0, len(in))
	for _, c := range in {
		if seen[c.guestID] {
			continue
		}
		seen[c.guestID] = true
		out = append(out, c)
	}
	return out
}
