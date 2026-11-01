package seed

import "context"

type campaignSpec struct {
	Key            string
	Name           string
	Description    string
	Tenant         string
	Segment        string
	AttributionDay int
	Status         string
	StartsDaysAgo  int
	EndsDaysAgo    int
}

func demoCampaigns() []campaignSpec {
	return []campaignSpec{
		{"camp-cinema", "Сеансы в кино", "SMS-кампания для гостей кинотеатра без кофейни", "cinema-sinema", "seg-sleep-cinema", 14, "active", 20, -10},
		{"camp-fit", "Первый визит в ФитЗону", "Приглашение на пробное занятие", "fit-zone", "seg-sleeping", 14, "finished", 60, 20},
		{"camp-kids", "Приветственная цепочка", "Email-цепочка для новых гостей детского центра", "kids-cube", "seg-newcomer", 7, "draft", 0, 0},
	}
}

func (s *Seeder) seedCampaigns(ctx context.Context) error {
	for _, c := range demoCampaigns() {
		id := demoID("campaign", c.Key)
		s.campaigns[c.Key] = id
		var startsAt, endsAt any
		if c.StartsDaysAgo > 0 {
			startsAt = s.now.AddDate(0, 0, -c.StartsDaysAgo)
		}
		if c.EndsDaysAgo != 0 {
			endsAt = s.now.AddDate(0, 0, -c.EndsDaysAgo)
		}
		if err := s.exec(ctx, `
			INSERT INTO campaign (id, name, description, target_tenant_id, segment_id, attribution_days,
				starts_at, ends_at, status, created_by, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (id) DO NOTHING`,
			id, c.Name, c.Description, s.tenants[c.Tenant], s.segments[c.Segment], c.AttributionDay,
			startsAt, endsAt, c.Status, s.users["marketer"], s.now.AddDate(0, 0, -c.StartsDaysAgo-1)); err != nil {
			return err
		}
	}
	return nil
}
