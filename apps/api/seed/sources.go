package seed

import (
	"context"
	"encoding/json"
)

type sourceSpec struct {
	Key          string
	Tenant       string
	TypeCode     string
	Name         string
	ScheduleCron string
	Active       bool
	FieldMapping map[string]string
}

func demoSources() []sourceSpec {
	return []sourceSpec{
		{"pos-bean", "cafe-bean", "pos", "POS Кофейня Bean", "", true, nil},
		{"crm-fit", "fit-zone", "crm", "CRM ФитЗона", "", true, nil},
		{"pos-cinema", "cinema-sinema", "pos", "POS Кинотеатр Синема", "", true, nil},
		{"crm-cinema", "cinema-sinema", "crm", "CRM Кинотеатр Синема", "", true, nil},
		{"csv-food", "food-plate", "csv_excel", "Выгрузка Ресторан Plate", "", true, nil},
		{"api-kids", "kids-cube", "api", "API Детский центр Куб", "0 */6 * * *", true, nil},
		{"pos-trend", "fashion-trend", "pos", "POS Магазин Тренд", "", true, nil},
		{"wifi-mall", "", "csv_excel", "Wi-Fi ТЦ", "", true, map[string]string{
			"external_id": "session_id",
			"occurred_at": "started_at",
			"event_type":  "event",
			"phone":       "msisdn",
			"tenant_code": "zone_code",
		}},
		{"pos-health", "health-plus", "pos", "POS Аптека Плюс", "", false, nil},
	}
}

func (s *Seeder) seedSources(ctx context.Context) error {
	for _, d := range demoSources() {
		typeID, err := s.typeID(ctx, d.TypeCode)
		if err != nil {
			return err
		}
		id := demoID("data_source", d.Key)
		s.sources[d.Key] = id
		var tenantID any
		if d.Tenant != "" {
			tenantID = s.tenants[d.Tenant]
		}
		var mapping any
		if len(d.FieldMapping) > 0 {
			raw, err := json.Marshal(d.FieldMapping)
			if err != nil {
				return err
			}
			mapping = string(raw)
		}
		var cron any
		if d.ScheduleCron != "" {
			cron = d.ScheduleCron
		}
		if err := s.exec(ctx, `
			INSERT INTO data_source (id, tenant_id, type_id, name, field_mapping, schedule_cron, is_active, last_import_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
			ON CONFLICT (id) DO NOTHING`,
			id, tenantID, typeID, d.Name, mapping, cron, d.Active, s.now.AddDate(0, 0, -2), s.now.AddDate(0, -6, 0)); err != nil {
			return err
		}
	}
	return nil
}
