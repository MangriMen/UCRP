package seed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/segments"
)

type segmentSpec struct {
	Key               string
	Name              string
	Description       string
	TemplateCode      string
	IsTemplate        bool
	Status            string
	VisitedTenantIDs  []string
	ExcludedTenantIDs []string
	PeriodDays        int
	Versions          []versionSpec
}

type versionSpec struct {
	No                int
	Status            string
	VisitedTenantIDs  []string
	ExcludedTenantIDs []string
	PeriodDays        int
	Members           bool
	DaysAgo           int
}

func demoSegments() []segmentSpec {
	return []segmentSpec{
		{
			Key: "tpl-sleeping", Name: "Спящий сосед", Description: "Гости, которые были в одном арендаторе, но не заходили в соседний",
			TemplateCode: "sleeping_neighbor", IsTemplate: true, Status: "active",
			VisitedTenantIDs: []string{"cinema-sinema"}, ExcludedTenantIDs: []string{"cafe-bean"}, PeriodDays: 90,
			Versions: []versionSpec{{No: 1, Status: "ready", VisitedTenantIDs: []string{"cinema-sinema"}, ExcludedTenantIDs: []string{"cafe-bean"}, PeriodDays: 90, Members: true, DaysAgo: 20}},
		},
		{
			Key: "tpl-single", Name: "Один магазин", Description: "Гости только одного арендатора за период",
			TemplateCode: "single_store", IsTemplate: true, Status: "active",
			VisitedTenantIDs: []string{"cafe-bean"}, ExcludedTenantIDs: []string{"cinema-sinema"}, PeriodDays: 60,
			Versions: []versionSpec{{No: 1, Status: "ready", VisitedTenantIDs: []string{"cafe-bean"}, ExcludedTenantIDs: []string{"cinema-sinema"}, PeriodDays: 60, Members: true, DaysAgo: 18}},
		},
		{
			Key: "tpl-churn", Name: "Риск оттока", Description: "Гости без визитов дольше 45 дней",
			TemplateCode: "churn_risk", IsTemplate: true, Status: "active",
			VisitedTenantIDs: []string{"fit-zone"}, ExcludedTenantIDs: []string{"food-plate"}, PeriodDays: 180,
			Versions: []versionSpec{{No: 1, Status: "ready", VisitedTenantIDs: []string{"fit-zone"}, ExcludedTenantIDs: []string{"food-plate"}, PeriodDays: 180, Members: true, DaysAgo: 15}},
		},
		{
			Key: "tpl-newcomer", Name: "Новичок", Description: "Гости с первым визитом в детский центр",
			TemplateCode: "newcomer", IsTemplate: true, Status: "active",
			VisitedTenantIDs: []string{"kids-cube"}, ExcludedTenantIDs: []string{"cinema-sinema"}, PeriodDays: 30,
			Versions: []versionSpec{{No: 1, Status: "ready", VisitedTenantIDs: []string{"kids-cube"}, ExcludedTenantIDs: []string{"cinema-sinema"}, PeriodDays: 30, Members: true, DaysAgo: 12}},
		},
		{
			Key: "tpl-high-value", Name: "Высокая ценность", Description: "Гости с покупками в нескольких арендаторах",
			TemplateCode: "high_value", IsTemplate: true, Status: "active",
			VisitedTenantIDs: []string{"cinema-sinema"}, ExcludedTenantIDs: []string{"pets-tail"}, PeriodDays: 120,
			Versions: []versionSpec{{No: 1, Status: "ready", VisitedTenantIDs: []string{"cinema-sinema"}, ExcludedTenantIDs: []string{"pets-tail"}, PeriodDays: 120, Members: true, DaysAgo: 10}},
		},
		{
			Key: "seg-sleep-cinema", Name: "Спящий сосед, кино", Description: "Кинотеатр без кофейни — кампания на кофе после сеанса",
			TemplateCode: "sleeping_neighbor", Status: "active",
			VisitedTenantIDs: []string{"cinema-sinema"}, ExcludedTenantIDs: []string{"cafe-bean"}, PeriodDays: 90,
			Versions: []versionSpec{
				{No: 1, Status: "ready", VisitedTenantIDs: []string{"cinema-sinema"}, ExcludedTenantIDs: []string{"cafe-bean"}, PeriodDays: 90, Members: true, DaysAgo: 25},
				{No: 2, Status: "ready", VisitedTenantIDs: []string{"cinema-sinema"}, ExcludedTenantIDs: []string{"cafe-bean"}, PeriodDays: 90, Members: true, DaysAgo: 5},
			},
		},
		{
			Key: "seg-single-archived", Name: "Один магазин (архив)", Description: "Архивная версия сегмента одного магазина",
			TemplateCode: "single_store", Status: "archived",
			VisitedTenantIDs: []string{"cafe-bean"}, ExcludedTenantIDs: []string{"cinema-sinema"}, PeriodDays: 60,
			Versions: []versionSpec{
				{No: 1, Status: "ready", VisitedTenantIDs: []string{"cafe-bean"}, ExcludedTenantIDs: []string{"cinema-sinema"}, PeriodDays: 60, Members: true, DaysAgo: 40},
				{No: 2, Status: "computing", VisitedTenantIDs: []string{"cafe-bean"}, ExcludedTenantIDs: []string{"cinema-sinema"}, PeriodDays: 60, DaysAgo: 2},
			},
		},
		{
			Key: "seg-failed", Name: "Экспериментальный сегмент", Description: "Сегмент с ошибкой расчёта",
			TemplateCode: "", Status: "active",
			VisitedTenantIDs: []string{"food-plate"}, ExcludedTenantIDs: []string{"cafe-bean"}, PeriodDays: 45,
			Versions: []versionSpec{
				{No: 1, Status: "failed", VisitedTenantIDs: []string{"food-plate"}, ExcludedTenantIDs: []string{"cafe-bean"}, PeriodDays: 45, DaysAgo: 3},
			},
		},
	}
}

func (s *Seeder) seedSegments(ctx context.Context) error {
	evaluator := segments.NewSqlEvaluator(s.db)
	for _, sg := range demoSegments() {
		segID := demoID("segment", sg.Key)
		s.segments[sg.Key] = segID
		createdAt := s.now.AddDate(0, 0, -30)
		var createdBy any
		if !sg.IsTemplate {
			createdBy = s.users["marketer"]
		}
		currentVersion := 1
		if len(sg.Versions) > 0 {
			currentVersion = sg.Versions[len(sg.Versions)-1].No
		}
		if err := s.exec(ctx, `
			INSERT INTO segment (id, name, description, template_code, is_template, current_version_no, status, created_by, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)
			ON CONFLICT (id) DO NOTHING`,
			segID, sg.Name, sg.Description, nullIfEmpty(sg.TemplateCode), sg.IsTemplate, currentVersion,
			sg.Status, createdBy, createdAt); err != nil {
			return err
		}
		for _, v := range sg.Versions {
			versionID := demoID("segment_version", fmt.Sprintf("%s/%d", sg.Key, v.No))
			s.versions[fmt.Sprintf("%s/%d", sg.Key, v.No)] = versionID
			rule := domain.SegmentRule{
				VisitedTenantIDs:  s.tenantIDs(v.VisitedTenantIDs),
				ExcludedTenantIDs: s.tenantIDs(v.ExcludedTenantIDs),
				PeriodDays:        v.PeriodDays,
			}
			ruleJSON, err := json.Marshal(rule)
			if err != nil {
				return err
			}
			versionCreatedAt := s.now.AddDate(0, 0, -v.DaysAgo)
			var memberCount, computedAt any
			if v.Status == "ready" {
				computedAt = versionCreatedAt
			}
			if err := s.exec(ctx, `
				INSERT INTO segment_version (id, segment_id, version_no, rule, status, member_count, computed_at, created_by, created_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
				ON CONFLICT (id) DO NOTHING`,
				versionID, segID, v.No, string(ruleJSON), v.Status, memberCount, computedAt, createdBy, versionCreatedAt); err != nil {
				return err
			}
			if !v.Members {
				continue
			}
			members, err := evaluator.Evaluate(ctx, rule)
			if err != nil {
				return fmt.Errorf("segment %s v%d: %w", sg.Key, v.No, err)
			}
			for _, guestID := range members {
				if err := s.exec(ctx, `
					INSERT INTO segment_member (segment_version_id, guest_id, added_at)
					VALUES ($1,$2,$3)
					ON CONFLICT DO NOTHING`, versionID, guestID, versionCreatedAt); err != nil {
					return err
				}
			}
			if err := s.exec(ctx, `
				UPDATE segment_version SET member_count = $2, computed_at = $3 WHERE id = $1`,
				versionID, len(members), versionCreatedAt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Seeder) tenantIDs(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		if id, ok := s.tenants[code]; ok {
			out = append(out, id)
		}
	}
	return out
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func (s *Seeder) segmentVersionID(key string, no int) string {
	return s.versions[fmt.Sprintf("%s/%d", key, no)]
}

func (s *Seeder) segmentCurrentVersion(key string) string {
	for _, sg := range demoSegments() {
		if sg.Key != key {
			continue
		}
		if len(sg.Versions) == 0 {
			return ""
		}
		return s.segmentVersionID(key, sg.Versions[len(sg.Versions)-1].No)
	}
	return ""
}
