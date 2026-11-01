package seed

import (
	"context"
	"fmt"

	"github.com/MangriMen/UCRP/internal/platform/auth"
)

type userSpec struct {
	Key              string
	Login            string
	DisplayName      string
	Role             string
	Tenant           string
	Active           bool
	CreatedDaysAgo   int
	LastLoginDaysAgo int
}

func demoUsers() []userSpec {
	return []userSpec{
		{"marketer", "marketer", "Мария Маркетолог", "marketer", "", true, 420, 1},
		{"manager", "manager", "Михаил Управляющий", "manager", "", true, 300, 3},
		{"analyst", "analyst", "Анна Аналитик", "analyst", "", true, 240, 0},
		{"tenant-cafe", "tenant.cafe", "Кофейня Bean", "tenant", "cafe-bean", true, 180, 2},
		{"tenant-cinema", "tenant.cinema", "Кинотеатр Синема", "tenant", "cinema-sinema", true, 150, 5},
		{"tenant-fit", "tenant.fit", "ФитЗона", "tenant", "fit-zone", true, 90, 1},
		{"analyst-old", "analyst.old", "Олег Бывший аналитик", "analyst", "", false, 500, 120},
	}
}

func demoUserEmail(login string) string {
	return fmt.Sprintf("%s@%s", login, demoMailDomain)
}

func (s *Seeder) seedUsers(ctx context.Context) error {
	hash, err := auth.HashPassword(demoPassword)
	if err != nil {
		return err
	}
	for _, u := range demoUsers() {
		id := demoID("app_user", u.Key)
		s.users[u.Key] = id
		var tenantID any
		if u.Tenant != "" {
			tenantID = s.tenants[u.Tenant]
		}
		if err := s.exec(ctx, `
			INSERT INTO app_user (id, email, password_hash, display_name, role, tenant_id, is_active, last_login_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING`,
			id, demoUserEmail(u.Login), hash, u.DisplayName, u.Role, tenantID, u.Active,
			s.now.AddDate(0, 0, -u.LastLoginDaysAgo), s.now.AddDate(0, 0, -u.CreatedDaysAgo)); err != nil {
			return err
		}
	}
	return nil
}

type tenantGrantSpec struct {
	UserKey string
	Tenant  string
}

var tenantGrantPermissions = []string{
	"reference.read",
	"tenant.aggregates.own",
	"metric.visits_count.read",
	"metric.new_guests_via_mall.read",
	"metric.repeat_visits.read",
}

func demoTenantGrants() []tenantGrantSpec {
	return []tenantGrantSpec{
		{"tenant-cafe", "cafe-bean"},
		{"tenant-cinema", "cinema-sinema"},
	}
}

func (s *Seeder) seedTenantGrants(ctx context.Context) error {
	for _, g := range demoTenantGrants() {
		id := demoID("access_grant", g.UserKey)
		tenantID := s.tenants[g.Tenant]
		if err := s.exec(ctx, `
			INSERT INTO access_grant (id, subject_type, subject_id, status, valid_from, created_by, reason, policy_version, created_at, updated_at)
			SELECT $1, 'tenant', $2, 'active', $3, $4, 'Демо: кабинет арендатора', version, $3, $3 FROM access_policy_state WHERE id = 1
			ON CONFLICT (id) DO NOTHING`,
			id, tenantID, s.now.AddDate(0, 0, -30), s.users["manager"]); err != nil {
			return err
		}
		for _, permission := range tenantGrantPermissions {
			if err := s.exec(ctx, `
				INSERT INTO access_grant_item (grant_id, permission_code, effect) VALUES ($1, $2, 'allow')
				ON CONFLICT DO NOTHING`, id, permission); err != nil {
				return err
			}
		}
		if err := s.exec(ctx, `
			INSERT INTO access_grant_scope (grant_id, scope_type, scope_value)
			VALUES ($1, 'tenant', $2), ($1, 'principal_role', 'tenant')
			ON CONFLICT DO NOTHING`, id, tenantID); err != nil {
			return err
		}
	}
	return s.exec(ctx, `UPDATE access_policy_state SET version = version + 1, updated_at = $1 WHERE id = 1`, s.now)
}
