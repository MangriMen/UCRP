package seed

import (
	"context"
	"encoding/json"
	"fmt"
)

type auditSpec struct {
	Key        string
	Action     string
	EntityType string
	EntityRef  string
	Actor      string
	Details    map[string]string
	DaysAgo    int
}

func demoAuditEntries() []auditSpec {
	return []auditSpec{
		{"a1", "merge.confirm", "merge_candidate", "mc-merged-3", "analyst", map[string]string{"score": "90"}, 70},
		{"a2", "merge.confirm", "merge_candidate", "mc-merged-1", "analyst", map[string]string{"score": "90"}, 88},
		{"a3", "merge.reject", "merge_candidate", "mc-merged-4", "analyst", map[string]string{"reason": "different_people"}, 60},
		{"a4", "merge.revert", "merge_candidate", "mc-merged-5", "analyst", map[string]string{"reason": "manual_unmerge"}, 45},
		{"a5", "export.create", "export_job", "exp-cinema-sms", "marketer", map[string]string{"channel": "sms"}, 20},
		{"a6", "export.create", "export_job", "exp-single-sms", "marketer", map[string]string{"channel": "sms"}, 18},
		{"a7", "export.create", "export_job", "exp-cinema-email", "marketer", map[string]string{"channel": "email"}, 15},
		{"a8", "export.create", "export_job", "exp-single-queued", "marketer", map[string]string{"channel": "sms"}, 1},
		{"a9", "export.create", "export_job", "exp-high-value-failed", "marketer", map[string]string{"channel": "push"}, 2},
		{"a10", "import.create", "import_batch", "bean-0901", "analyst", map[string]string{"mode": "file"}, 60},
		{"a11", "import.create", "import_batch", "fit-0901", "analyst", map[string]string{"mode": "file"}, 60},
		{"a12", "import.create", "import_batch", "cinema-1001", "analyst", map[string]string{"mode": "file"}, 30},
		{"a13", "import.create", "import_batch", "food-0920", "analyst", map[string]string{"mode": "file"}, 41},
		{"a14", "import.create", "import_batch", "wifi-1010", "analyst", map[string]string{"mode": "file"}, 21},
		{"a15", "import.create", "import_batch", "kids-pull-1025", "analyst", map[string]string{"mode": "pull"}, 6},
		{"a16", "segment.create", "segment", "seg-sleep-cinema", "marketer", map[string]string{"template": "sleeping_neighbor"}, 25},
		{"a17", "segment.create", "segment", "tpl-single", "marketer", map[string]string{"template": "single_store"}, 18},
		{"a18", "segment.create", "segment", "seg-failed", "marketer", map[string]string{"template": ""}, 3},
		{"a19", "data_source.create", "data_source", "pos-bean", "analyst", map[string]string{"type": "pos"}, 180},
		{"a20", "data_source.create", "data_source", "wifi-mall", "analyst", map[string]string{"type": "csv_excel"}, 175},
	}
}

func (s *Seeder) seedAudit(ctx context.Context) error {
	for _, a := range demoAuditEntries() {
		entityID := s.auditEntityID(a.EntityType, a.EntityRef)
		if entityID == "" {
			return fmt.Errorf("audit %s: unknown entity %s/%s", a.Key, a.EntityType, a.EntityRef)
		}
		details, err := json.Marshal(a.Details)
		if err != nil {
			return err
		}
		if err := s.exec(ctx, `
			INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, request_id, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (id) DO NOTHING`,
			demoID("audit_log", a.Key), s.users[a.Actor], actorRole(a.Actor), a.Action, a.EntityType, entityID,
			string(details), "demo-"+a.Key, s.now.AddDate(0, 0, -a.DaysAgo)); err != nil {
			return err
		}
	}
	return nil
}

func actorRole(userKey string) string {
	for _, u := range demoUsers() {
		if u.Key == userKey {
			return u.Role
		}
	}
	return "analyst"
}

func (s *Seeder) auditEntityID(entityType, ref string) string {
	switch entityType {
	case "merge_candidate":
		return demoID("merge_candidate", ref)
	case "export_job":
		if id, ok := s.exports[ref]; ok {
			return id
		}
		return demoID("export_job", ref)
	case "import_batch":
		return s.batches[ref]
	case "segment":
		return s.segments[ref]
	case "data_source":
		return s.sources[ref]
	default:
		return ""
	}
}
