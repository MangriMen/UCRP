package seed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type categorySpec struct {
	Code string
	Name string
}

type relationSpec struct {
	From   string
	To     string
	Type   string
	Weight int
}

type tenantSpec struct {
	Code     string
	Name     string
	Category string
	Kind     string
	Floor    int
	X        int
	Y        int
	Priority int
	Active   bool
	Deleted  bool
}

type projectSpec struct {
	Tenant string
	Code   string
	Name   string
}

func demoCategories() []categorySpec {
	return []categorySpec{
		{"cafe", "Кофейни и пекарни"},
		{"food", "Рестораны и фудкорт"},
		{"sport", "Спорт и фитнес"},
		{"cinema", "Кинотеатры"},
		{"kids", "Детские центры"},
		{"beauty", "Красота и уход"},
		{"fashion", "Одежда и обувь"},
		{"electronics", "Электроника"},
		{"health", "Аптеки и здоровье"},
		{"books", "Книги и канцелярия"},
		{"pets", "Зоотовары"},
		{"services", "Бытовые услуги"},
	}
}

func demoRelations() []relationSpec {
	return []relationSpec{
		{"cafe", "cinema", "adjacent", 85},
		{"food", "cinema", "adjacent", 80},
		{"sport", "health", "adjacent", 70},
		{"sport", "food", "adjacent", 65},
		{"kids", "fashion", "adjacent", 60},
		{"books", "kids", "adjacent", 75},
		{"pets", "health", "adjacent", 55},
		{"beauty", "fashion", "adjacent", 90},
		{"cafe", "books", "adjacent", 70},
		{"electronics", "services", "adjacent", 60},
		{"cafe", "food", "competitor", 100},
		{"cinema", "kids", "competitor", 95},
		{"sport", "beauty", "competitor", 90},
		{"fashion", "electronics", "competitor", 85},
		{"health", "beauty", "competitor", 80},
		{"books", "electronics", "competitor", 70},
	}
}

func demoTenants() []tenantSpec {
	return []tenantSpec{
		{"cafe-bean", "Кофейня Bean", "cafe", "mall", 1, 120, 80, 70, true, false},
		{"fit-zone", "ФитЗона", "sport", "mall", 2, 340, 210, 85, true, false},
		{"cinema-sinema", "Кинотеатр Синема", "cinema", "mall", 3, 500, 120, 90, true, false},
		{"food-plate", "Ресторан Plate", "food", "mall", 2, 210, 300, 75, true, false},
		{"kids-cube", "Детский центр Куб", "kids", "mall", 3, 620, 340, 60, true, false},
		{"beauty-lokon", "Салон Локон", "beauty", "mall", 1, 90, 260, 55, true, false},
		{"fashion-trend", "Магазин Тренд", "fashion", "mall", 1, 60, 400, 65, true, false},
		{"tech-volt", "Техника Вольт", "electronics", "mall", 2, 700, 180, 50, true, false},
		{"health-plus", "Аптека Плюс", "health", "mall", 1, 30, 120, 40, false, false},
		{"food-bread", "Пекарня Хлеб", "food", "mall", 1, 150, 60, 45, true, false},
		{"sport-protein", "Спортпит Протеин", "sport", "mall", 2, 380, 260, 35, true, false},
		{"books-line", "Книжный Линия", "books", "mall", 3, 560, 420, 30, true, false},
		{"pets-tail", "Зоомагазин Хвост", "pets", "mall", 1, 40, 480, 25, true, false},
		{"svc-parking", "Паркинг ТЦ", "services", "mall", 0, 0, 0, 20, true, false},
		{"spa-aqua", "Спа Аква", "beauty", "external", 0, 0, 0, 15, true, false},
		{"cinema-lux", "Кинотеатр Люкс", "cinema", "external", 0, 0, 0, 10, true, true},
	}
}

func demoProjects() []projectSpec {
	return []projectSpec{
		{"fit-zone", "gym", "Тренажёрный зал"},
		{"fit-zone", "pool", "Бассейн"},
		{"cinema-sinema", "imax", "Зал IMAX"},
		{"cinema-sinema", "vip", "VIP-зал"},
		{"cinema-sinema", "kids-hall", "Детский зал"},
		{"food-plate", "main", "Основной зал"},
		{"food-plate", "banquet", "Банкетный зал"},
		{"kids-cube", "main", "Основная площадка"},
	}
}

func (s *Seeder) seedReference(ctx context.Context) error {
	if err := s.ensureEventTypes(ctx); err != nil {
		return err
	}
	if err := s.ensureDataSourceTypes(ctx); err != nil {
		return err
	}
	for _, c := range demoCategories() {
		id := demoID("tenant_category", c.Code)
		s.categories[c.Code] = id
		if err := s.exec(ctx, `
			INSERT INTO tenant_category (id, code, name, is_active, created_at)
			VALUES ($1, $2, $3, TRUE, $4)
			ON CONFLICT (id) DO NOTHING`, id, c.Code, c.Name, s.now); err != nil {
			return err
		}
	}
	for _, r := range demoRelations() {
		if err := s.exec(ctx, `
			INSERT INTO tenant_category_relation (category_id, related_category_id, relation_type, weight_pct)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (category_id, related_category_id, relation_type) DO NOTHING`,
			s.categories[r.From], s.categories[r.To], r.Type, r.Weight); err != nil {
			return err
		}
	}
	for _, t := range demoTenants() {
		id := demoID("tenant", t.Code)
		s.tenants[t.Code] = id
		var deletedAt any
		if t.Deleted {
			deletedAt = s.now.AddDate(0, -2, 0)
		}
		if err := s.exec(ctx, `
			INSERT INTO tenant (id, code, name, category_id, kind, floor, x_m, y_m, priority, is_active, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11, $12)
			ON CONFLICT (id) DO NOTHING`,
			id, t.Code, t.Name, s.categories[t.Category], t.Kind, t.Floor, t.X, t.Y, t.Priority, t.Active, s.now, deletedAt); err != nil {
			return err
		}
	}
	for _, p := range demoProjects() {
		key := p.Tenant + "/" + p.Code
		id := demoID("project", key)
		s.projects[key] = id
		if err := s.exec(ctx, `
			INSERT INTO project (id, tenant_id, code, name, is_active)
			VALUES ($1, $2, $3, $4, TRUE)
			ON CONFLICT (id) DO NOTHING`, id, s.tenants[p.Tenant], p.Code, p.Name); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) ensureEventTypes(ctx context.Context) error {
	defaults := []struct {
		code           string
		name           string
		hasAmount      bool
		countsPresence bool
	}{
		{"visit", "Визит", false, true},
		{"purchase", "Покупка", true, true},
		{"booking", "Бронь", true, false},
		{"top_up", "Пополнение", true, false},
		{"usage", "Списание", true, false},
		{"reversal", "Возврат/отмена", true, false},
	}
	for _, d := range defaults {
		var id string
		err := s.db.QueryRowContext(ctx, `SELECT id FROM event_type WHERE code = $1`, d.code).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			id = demoID("event_type", d.code)
			if err := s.exec(ctx, `
				INSERT INTO event_type (id, code, name, has_amount, counts_as_presence, is_active)
				VALUES ($1, $2, $3, $4, $5, TRUE)`, id, d.code, d.name, d.hasAmount, d.countsPresence); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		s.eventTypes[d.code] = id
	}
	return s.exec(ctx, `
		UPDATE event_type SET reverses_event_type_id = $1 WHERE code = 'reversal' AND reverses_event_type_id IS NULL`,
		s.eventTypes["purchase"])
}

func (s *Seeder) ensureDataSourceTypes(ctx context.Context) error {
	defaults := []struct {
		code         string
		name         string
		connectorKey string
		supportsFile bool
		supportsPull bool
	}{
		{"csv_excel", "CSV/Excel", "csv", true, false},
		{"crm", "CRM", "csv", true, false},
		{"pos", "Касса/POS", "csv", true, false},
		{"api", "API", "api", false, true},
	}
	for _, d := range defaults {
		var exists int
		err := s.db.QueryRowContext(ctx, `SELECT 1 FROM data_source_type WHERE code = $1`, d.code).Scan(&exists)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err := s.exec(ctx, `
			INSERT INTO data_source_type (id, code, name, connector_key, supports_file, supports_pull, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, TRUE)`,
			demoID("data_source_type", d.code), d.code, d.name, d.connectorKey, d.supportsFile, d.supportsPull); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) typeID(ctx context.Context, code string) (string, error) {
	var id string
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM data_source_type WHERE code = $1`, code).Scan(&id); err != nil {
		return "", fmt.Errorf("data source type %s: %w", code, err)
	}
	return id, nil
}
