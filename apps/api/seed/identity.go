package seed

import (
	"context"
	"fmt"
	"time"
)

type candidateSpec struct {
	Key        string
	GuestA     string
	GuestB     string
	MatchType  string
	ReasonCode string
	ReasonText string
	Score      int
	Resolution string
	Winner     string
	Loser      string
	DecidedBy  string
	DaysAgo    int
}

type blockSpec struct {
	Key       string
	GuestA    string
	GuestB    string
	Reason    string
	Candidate string
	DaysAgo   int
}

func demoMergeCandidates() []candidateSpec {
	return []candidateSpec{
		{"mc-olga-igor", "g-olga", "g-igor", "phone", "same_phone", "Совпадение нормализованного телефона", 90, "pending", "", "", "", 20},
		{"mc-anna-email", "g-anna", "g-email", "email", "same_email", "Совпадение email в разных источниках", 70, "pending", "", "", "", 18},
		{"mc-card-vip", "g-card", "g-vip", "card", "same_card", "Совпадение номера карты лояльности", 75, "pending", "", "", "", 15},
		{"mc-merged-1", "bg-003", "g-merged-1", "phone", "same_phone", "Автоматическое объединение по телефону", 90, "auto_merged", "bg-003", "g-merged-1", "analyst", 88},
		{"mc-merged-2", "bg-006", "g-merged-2", "phone", "same_phone", "Автоматическое объединение по телефону", 90, "auto_merged", "bg-006", "g-merged-2", "analyst", 86},
		{"mc-merged-3", "bg-009", "g-merged-3", "phone", "same_phone", "Подтверждено аналитиком после проверки", 90, "confirmed", "bg-009", "g-merged-3", "analyst", 70},
		{"mc-merged-4", "bg-012", "g-merged-4", "email", "same_email", "Отклонено: разные люди с общим семейным адресом", 70, "rejected", "", "", "analyst", 60},
		{"mc-merged-5", "bg-015", "g-merged-5", "phone", "same_phone", "Объединение отменено вручную", 90, "reverted", "bg-015", "g-merged-5", "analyst", 45},
	}
}

func demoUnmergeBlocks() []blockSpec {
	return []blockSpec{
		{"ub-1", "bg-015", "g-merged-5", "manual_unmerge", "mc-merged-5", 44},
		{"ub-2", "bg-012", "g-merged-4", "rejected", "mc-merged-4", 59},
		{"ub-3", "bg-018", "g-merged-6", "manual_unmerge", "", 30},
	}
}

func (s *Seeder) seedIdentity(ctx context.Context) error {
	for _, c := range demoMergeCandidates() {
		a, b := s.guests[c.GuestA], s.guests[c.GuestB]
		if a == "" || b == "" {
			return fmt.Errorf("merge candidate %s: unknown guest", c.Key)
		}
		if a > b {
			a, b = b, a
		}
		createdAt := s.now.AddDate(0, 0, -c.DaysAgo)
		var winner, loser, decidedBy, decidedAt any
		if c.Winner != "" {
			winner = s.guests[c.Winner]
		}
		if c.Loser != "" {
			loser = s.guests[c.Loser]
		}
		if c.DecidedBy != "" {
			decidedBy = s.users[c.DecidedBy]
			decidedAt = createdAt.Add(2 * time.Hour)
		}
		if err := s.exec(ctx, `
			INSERT INTO merge_candidate (id, guest_a_id, guest_b_id, match_type, reason_code, reason_text, score,
				resolution, winner_guest_id, merged_guest_id, decided_by, decided_at, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (id) DO NOTHING`,
			demoID("merge_candidate", c.Key), a, b, c.MatchType, c.ReasonCode, c.ReasonText, c.Score,
			c.Resolution, winner, loser, decidedBy, decidedAt, createdAt); err != nil {
			return err
		}
	}
	for _, b := range demoUnmergeBlocks() {
		a, other := s.guests[b.GuestA], s.guests[b.GuestB]
		if a == "" || other == "" {
			return fmt.Errorf("unmerge block %s: unknown guest", b.Key)
		}
		if a > other {
			a, other = other, a
		}
		var candidateID any
		if b.Candidate != "" {
			candidateID = demoID("merge_candidate", b.Candidate)
		}
		if err := s.exec(ctx, `
			INSERT INTO unmerge_block (id, guest_a_id, guest_b_id, reason, candidate_id, created_by, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (id) DO NOTHING`,
			demoID("unmerge_block", b.Key), a, other, b.Reason, candidateID, s.users["analyst"],
			s.now.AddDate(0, 0, -b.DaysAgo)); err != nil {
			return err
		}
	}
	return nil
}
