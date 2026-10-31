package identity

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

const resolveBatchLimit = 5000

var reliableKeyOrder = []string{"phone", "email", "card"}

type Resolver struct {
	guests      domain.GuestRepo
	identifiers domain.GuestIdentifierRepo
	events      domain.EventRepo
	candidates  domain.MergeCandidateRepo
	blocks      domain.UnmergeBlockRepo
	clock       domain.Clock
	idGen       domain.IDGen
}

func NewResolver(
	guests domain.GuestRepo,
	identifiers domain.GuestIdentifierRepo,
	events domain.EventRepo,
	candidates domain.MergeCandidateRepo,
	blocks domain.UnmergeBlockRepo,
) *Resolver {
	return &Resolver{
		guests:      guests,
		identifiers: identifiers,
		events:      events,
		candidates:  candidates,
		blocks:      blocks,
		clock:       domain.SystemClock{},
		idGen:       idgen.UUID{},
	}
}

func (r *Resolver) ResolveBatch(ctx context.Context, batchID string) error {
	events, err := r.events.ListUnmatchedByBatch(ctx, batchID, resolveBatchLimit)
	if err != nil {
		return err
	}
	for i := range events {
		if err := r.resolveEvent(ctx, &events[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) resolveEvent(ctx context.Context, e *domain.Event) error {
	keys := parseMatchKeys(e.MatchKeys)
	if len(keys) == 0 {
		return nil
	}
	guestIDs, err := r.lookupGuests(ctx, keys)
	if err != nil {
		return err
	}
	switch len(guestIDs) {
	case 0:
		return r.createAndAttach(ctx, e, keys)
	case 1:
		return r.attach(ctx, e, guestIDs[0], keys)
	default:
		return r.resolveConflict(ctx, e, guestIDs, keys)
	}
}

func (r *Resolver) lookupGuests(ctx context.Context, keys map[string]string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, 2)
	for _, idType := range reliableKeyOrder {
		value := keys[idType]
		if value == "" {
			continue
		}
		ids, err := r.identifiers.FindByValue(ctx, idType, value)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			active, err := r.activeGuest(ctx, id.GuestID)
			if err != nil {
				return nil, err
			}
			if active == "" {
				continue
			}
			if _, dup := seen[active]; dup {
				continue
			}
			seen[active] = struct{}{}
			out = append(out, active)
		}
	}
	return out, nil
}

func (r *Resolver) activeGuest(ctx context.Context, guestID string) (string, error) {
	g, err := r.guests.GetByID(ctx, guestID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	switch g.Status {
	case domain.GuestActive:
		return g.ID, nil
	case domain.GuestMerged:
		if g.MergedIntoID == "" {
			return "", nil
		}
		return r.activeGuest(ctx, g.MergedIntoID)
	default:
		return "", nil
	}
}

func (r *Resolver) createAndAttach(ctx context.Context, e *domain.Event, keys map[string]string) error {
	now := r.clock.Now()
	guestID := r.idGen.NewID(ctx)
	guest := &domain.Guest{
		ID:        guestID,
		Status:    domain.GuestActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.guests.Insert(ctx, guest); err != nil {
		return err
	}
	if err := r.insertIdentifiers(ctx, guestID, e.DataSourceID, keys, now); err != nil {
		return err
	}
	return r.events.AttachGuest(ctx, e.ID, guestID)
}

func (r *Resolver) attach(ctx context.Context, e *domain.Event, guestID string, keys map[string]string) error {
	if err := r.insertIdentifiers(ctx, guestID, e.DataSourceID, keys, r.clock.Now()); err != nil {
		return err
	}
	return r.events.AttachGuest(ctx, e.ID, guestID)
}

type pair struct {
	winner     string
	loser      string
	matchType  string
	score      int
	auto       bool
	reasonCode string
	reasonText string
}

func (r *Resolver) resolveConflict(ctx context.Context, e *domain.Event, guestIDs []string, keys map[string]string) error {
	now := r.clock.Now()
	p := r.pairOf(ctx, guestIDs, keys)
	blocked, err := r.blocks.HasBlock(ctx, p.winner, p.loser)
	if err != nil {
		return err
	}
	if !blocked {
		if p.auto {
			_, err = r.candidates.AutoMerge(ctx, p.winner, p.loser, p.matchType, p.reasonCode, p.reasonText, p.score, now)
		} else {
			_, err = r.candidates.CreatePending(ctx, p.winner, p.loser, p.matchType, p.reasonCode, p.reasonText, p.score, now)
		}
		if err != nil {
			return err
		}
	}
	return r.attach(ctx, e, p.winner, keys)
}

func (r *Resolver) pairOf(ctx context.Context, guestIDs []string, keys map[string]string) pair {
	p := pair{
		winner:     guestIDs[0],
		loser:      guestIDs[1],
		matchType:  "keys",
		score:      50,
		reasonCode: "multi_key_conflict",
		reasonText: "Ключи события указывают на разных гостей",
	}
	for _, idType := range reliableKeyOrder {
		value := keys[idType]
		if value == "" {
			continue
		}
		ids, err := r.identifiers.FindByValue(ctx, idType, value)
		if err != nil {
			continue
		}
		owners := map[string]bool{}
		for _, id := range ids {
			owners[id.GuestID] = true
		}
		if owners[p.winner] && owners[p.loser] {
			p.matchType = idType
			break
		}
	}
	switch p.matchType {
	case "phone":
		p.score, p.auto = 90, true
		p.reasonCode, p.reasonText = "shared_phone", "Общий телефон"
	case "email":
		p.score = 70
		p.reasonCode, p.reasonText = "shared_email", "Общий email"
	case "card":
		p.score = 75
		p.reasonCode, p.reasonText = "shared_card", "Общая карта"
	}
	return p
}

func (r *Resolver) insertIdentifiers(ctx context.Context, guestID, dataSourceID string, keys map[string]string, now time.Time) error {
	for _, idType := range reliableKeyOrder {
		value := keys[idType]
		if value == "" {
			continue
		}
		existing, err := r.identifiers.FindByValue(ctx, idType, value)
		if err != nil {
			return err
		}
		if ownsIdentifier(existing, guestID) {
			continue
		}
		gi := &domain.GuestIdentifier{
			ID:              r.idGen.NewID(ctx),
			GuestID:         guestID,
			OriginGuestID:   guestID,
			IDType:          idType,
			ValueNormalized: value,
			DataSourceID:    dataSourceID,
			FirstSeenAt:     now,
		}
		if err := r.identifiers.Insert(ctx, gi); err != nil && !isDuplicateKey(err) {
			return err
		}
	}
	return nil
}

func ownsIdentifier(ids []domain.GuestIdentifier, guestID string) bool {
	for _, id := range ids {
		if id.GuestID == guestID {
			return true
		}
	}
	return false
}

func parseMatchKeys(raw string) map[string]string {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint")
}
