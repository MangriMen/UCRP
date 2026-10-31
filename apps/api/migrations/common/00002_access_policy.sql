-- +goose Up

CREATE TABLE permission (
  code          VARCHAR(64) PRIMARY KEY,
  domain        VARCHAR(32) NOT NULL,
  kind          VARCHAR(16) NOT NULL CHECK (kind IN ('action', 'metric', 'field')),
  title         VARCHAR(200) NOT NULL,
  description   TEXT,
  is_sensitive  BOOLEAN NOT NULL DEFAULT FALSE,
  min_k         SMALLINT CHECK (min_k IS NULL OR min_k >= 0),
  is_active     BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX ix_permission_domain ON permission(domain, kind);

INSERT INTO permission (code, domain, kind, title, is_sensitive, min_k) VALUES
  ('guest.profile.read', 'guest', 'action', 'Профиль гостя', TRUE, NULL),
  ('guest.profile.search', 'guest', 'action', 'Поиск гостей', TRUE, NULL),
  ('guest.profile.delete', 'guest', 'action', 'Soft-delete гостя', TRUE, NULL),
  ('guest.event.read', 'guest', 'action', 'Лента событий', TRUE, NULL),
  ('guest.identifier.read', 'guest', 'action', 'Идентификаторы (телефон, email, карта)', TRUE, NULL),
  ('guest.contact_history.read', 'guest', 'action', 'История контактов', TRUE, NULL),
  ('segment.read', 'segment', 'action', 'Список и карточка сегмента', FALSE, NULL),
  ('segment.preview', 'segment', 'action', 'Предпросмотр правила', FALSE, NULL),
  ('segment.write', 'segment', 'action', 'Создать/изменить сегмент', FALSE, NULL),
  ('segment.delete', 'segment', 'action', 'Архивировать сегмент', FALSE, NULL),
  ('segment.members.read', 'segment', 'action', 'Состав сегмента (строки гостей)', TRUE, NULL),
  ('recommendation.read', 'recommendation', 'action', 'Рекомендации гостя', TRUE, NULL),
  ('recommendation.refresh', 'recommendation', 'action', 'Пересчёт рекомендаций', FALSE, NULL),
  ('identity.queue.read', 'identity', 'action', 'Очередь спорных склеек', TRUE, NULL),
  ('identity.decide', 'identity', 'action', 'Confirm / reject / revert', TRUE, NULL),
  ('import.read', 'import', 'action', 'Журнал загрузок', FALSE, NULL),
  ('import.create', 'import', 'action', 'Загрузка файла, ручная синхронизация', FALSE, NULL),
  ('import.rejected_rows.read', 'import', 'action', 'Отклонённые строки', FALSE, NULL),
  ('data_source.read', 'data_source', 'action', 'Источники данных', FALSE, NULL),
  ('data_source.write', 'data_source', 'action', 'Создать/изменить источник', FALSE, NULL),
  ('reference.read', 'reference', 'action', 'Справочники (арендаторы, категории, проекты, типы)', FALSE, NULL),
  ('reference.write', 'reference', 'action', 'Изменение справочников', FALSE, NULL),
  ('export.read', 'export', 'action', 'Журнал выгрузок', FALSE, NULL),
  ('export.create', 'export', 'action', 'Создать выгрузку', TRUE, NULL),
  ('export.download', 'export', 'action', 'Скачать CSV получателей', TRUE, NULL),
  ('export.members.read', 'export', 'action', 'Получатели и контрольная группа', TRUE, NULL),
  ('campaign.read', 'campaign', 'action', 'Кампании', FALSE, NULL),
  ('campaign.write', 'campaign', 'action', 'Создать/изменить кампанию', FALSE, NULL),
  ('campaign.results.read', 'campaign', 'action', 'Результаты кампании', FALSE, NULL),
  ('contact_limit.read', 'contact_limit', 'action', 'Правила лимита контактов', FALSE, NULL),
  ('contact_limit.write', 'contact_limit', 'action', 'Изменить лимит', FALSE, NULL),
  ('user.read', 'user', 'action', 'Пользователи', FALSE, NULL),
  ('user.write', 'user', 'action', 'Создать/изменить/деактивировать', TRUE, NULL),
  ('audit.read', 'audit', 'action', 'Журнал аудита', TRUE, NULL),
  ('access.grant.read', 'access', 'action', 'Просмотр грантов', FALSE, NULL),
  ('access.grant.write', 'access', 'action', 'Выдача и отзыв грантов', TRUE, NULL),
  ('access.simulate', 'access', 'action', '«Что увидит арендатор» без выдачи', FALSE, NULL),
  ('data_quality.read', 'data_quality', 'action', 'Качество источников', FALSE, NULL),
  ('dashboard.read', 'dashboard', 'action', 'Дашборд (каркас, без метрик)', FALSE, NULL),
  ('tenant.aggregates.read', 'tenant', 'action', 'Агрегаты любого арендатора', FALSE, NULL),
  ('tenant.aggregates.own', 'tenant', 'action', 'Агрегаты только своего арендатора', FALSE, NULL),
  ('metric.first_purchase_new_tenant.read', 'metric', 'metric', 'Первая покупка у нового арендатора (главная, A-17)', TRUE, 5),
  ('metric.cross_tenant_distribution.read', 'metric', 'metric', 'Cross-tenant: распределение гостей по числу арендаторов (1 / 2 / 3+)', TRUE, 5),
  ('metric.cross_tenant_rate.read', 'metric', 'metric', 'Доля гостей с ≥ 2 арендаторами', TRUE, 5),
  ('metric.match_rate.read', 'metric', 'metric', 'Match rate (≥ 70%)', FALSE, 5),
  ('metric.disputed_merge_rate.read', 'metric', 'metric', 'Доля оспоренных склеек (< 1%)', FALSE, 5),
  ('metric.segment_to_export_time.read', 'metric', 'metric', 'Время «сегмент → выгрузка» (медиана, p95)', FALSE, 5),
  ('metric.conversion_rate.read', 'metric', 'metric', 'Конверсия рассылки (≥ 3%)', TRUE, 5),
  ('metric.conversion_lift.read', 'metric', 'metric', 'Инкремент против контрольной группы', TRUE, 5),
  ('metric.new_guests_via_mall.read', 'metric', 'metric', 'Новые гости, пришедшие через комплекс', TRUE, 5),
  ('metric.repeat_visits.read', 'metric', 'metric', 'Повторные визиты', TRUE, 5),
  ('metric.revenue_total.read', 'metric', 'metric', 'Суммарная выручка по гостю/периоду', TRUE, 5),
  ('metric.avg_check.read', 'metric', 'metric', 'Средний чек', TRUE, 5),
  ('metric.visits_count.read', 'metric', 'metric', 'Число визитов', FALSE, 5),
  ('metric.segment_size.read', 'metric', 'metric', 'Размер сегмента', FALSE, 5),
  ('metric.export_counts.read', 'metric', 'metric', 'Счётчики выгрузки', FALSE, 5),
  ('metric.contact_frequency.read', 'metric', 'metric', 'Частота контактов', FALSE, 5),
  ('metric.data_quality.read', 'metric', 'metric', 'Доля отклонённых строк, свежесть', FALSE, 5),
  ('metric.merge_queue_size.read', 'metric', 'metric', 'Размер очереди спорных склеек', FALSE, 5),
  ('metric.campaign_results.read', 'metric', 'metric', 'Конверсия и инкремент по кампании', TRUE, 5),
  ('metric.tenant_priority.read', 'metric', 'metric', 'Приоритет продвижения арендатора', TRUE, 5),
  ('field.guest.display_name.read', 'field', 'field', 'displayName', TRUE, NULL),
  ('field.guest.phone.read', 'field', 'field', 'identifiers[].valueNormalized для phone', TRUE, NULL),
  ('field.guest.email.read', 'field', 'field', 'identifiers[].valueNormalized для email', TRUE, NULL),
  ('field.guest.card.read', 'field', 'field', 'identifiers[].valueNormalized для card', TRUE, NULL),
  ('field.guest.total_amount.read', 'field', 'field', 'totalAmountMinor, currency', TRUE, NULL),
  ('field.guest.identifiers.read', 'field', 'field', 'Массив identifiers целиком', TRUE, NULL),
  ('field.guest.segments.read', 'field', 'field', 'Массив segments', FALSE, NULL),
  ('field.guest.tenants_visited.read', 'field', 'field', 'Массив tenants, tenantCount', TRUE, NULL),
  ('field.event.amount.read', 'field', 'field', 'event.amountMinor', TRUE, NULL),
  ('field.event.raw_payload.read', 'field', 'field', 'event.rawPayload', TRUE, NULL),
  ('field.event.external_id.read', 'field', 'field', 'event.externalId', FALSE, NULL),
  ('field.export.contact.read', 'field', 'field', 'Контакт в выгрузке', TRUE, NULL),
  ('field.tenant.priority.read', 'field', 'field', 'tenant.priority', FALSE, NULL);

CREATE TABLE role_permission (
  role            VARCHAR(16) NOT NULL CHECK (role IN ('marketer', 'manager', 'tenant', 'analyst', 'system')),
  permission_code VARCHAR(64) NOT NULL REFERENCES permission(code),
  effect          VARCHAR(8) NOT NULL CHECK (effect IN ('allow', 'deny')),
  PRIMARY KEY (role, permission_code)
);

INSERT INTO role_permission (role, permission_code, effect) VALUES
  ('marketer', 'guest.profile.read', 'allow'),
  ('analyst', 'guest.profile.read', 'allow'),
  ('marketer', 'guest.profile.search', 'allow'),
  ('analyst', 'guest.profile.search', 'allow'),
  ('analyst', 'guest.profile.delete', 'allow'),
  ('marketer', 'guest.event.read', 'allow'),
  ('analyst', 'guest.event.read', 'allow'),
  ('marketer', 'guest.identifier.read', 'allow'),
  ('analyst', 'guest.identifier.read', 'allow'),
  ('marketer', 'guest.contact_history.read', 'allow'),
  ('analyst', 'guest.contact_history.read', 'allow'),
  ('marketer', 'segment.read', 'allow'),
  ('manager', 'segment.read', 'allow'),
  ('analyst', 'segment.read', 'allow'),
  ('marketer', 'segment.preview', 'allow'),
  ('analyst', 'segment.preview', 'allow'),
  ('marketer', 'segment.write', 'allow'),
  ('analyst', 'segment.write', 'allow'),
  ('marketer', 'segment.delete', 'allow'),
  ('analyst', 'segment.delete', 'allow'),
  ('marketer', 'segment.members.read', 'allow'),
  ('analyst', 'segment.members.read', 'allow'),
  ('marketer', 'recommendation.read', 'allow'),
  ('analyst', 'recommendation.read', 'allow'),
  ('marketer', 'recommendation.refresh', 'allow'),
  ('analyst', 'recommendation.refresh', 'allow'),
  ('marketer', 'identity.queue.read', 'allow'),
  ('analyst', 'identity.queue.read', 'allow'),
  ('analyst', 'identity.decide', 'allow'),
  ('manager', 'import.read', 'allow'),
  ('analyst', 'import.read', 'allow'),
  ('analyst', 'import.create', 'allow'),
  ('analyst', 'import.rejected_rows.read', 'allow'),
  ('marketer', 'data_source.read', 'allow'),
  ('manager', 'data_source.read', 'allow'),
  ('analyst', 'data_source.read', 'allow'),
  ('analyst', 'data_source.write', 'allow'),
  ('marketer', 'reference.read', 'allow'),
  ('manager', 'reference.read', 'allow'),
  ('tenant', 'reference.read', 'allow'),
  ('analyst', 'reference.read', 'allow'),
  ('analyst', 'reference.write', 'allow'),
  ('marketer', 'export.read', 'allow'),
  ('manager', 'export.read', 'allow'),
  ('analyst', 'export.read', 'allow'),
  ('marketer', 'export.create', 'allow'),
  ('marketer', 'export.download', 'allow'),
  ('marketer', 'export.members.read', 'allow'),
  ('marketer', 'campaign.read', 'allow'),
  ('manager', 'campaign.read', 'allow'),
  ('analyst', 'campaign.read', 'allow'),
  ('marketer', 'campaign.write', 'allow'),
  ('marketer', 'campaign.results.read', 'allow'),
  ('manager', 'campaign.results.read', 'allow'),
  ('marketer', 'contact_limit.read', 'allow'),
  ('manager', 'contact_limit.read', 'allow'),
  ('analyst', 'contact_limit.read', 'allow'),
  ('marketer', 'contact_limit.write', 'allow'),
  ('manager', 'user.read', 'allow'),
  ('manager', 'user.write', 'allow'),
  ('manager', 'audit.read', 'allow'),
  ('analyst', 'audit.read', 'allow'),
  ('manager', 'access.grant.read', 'allow'),
  ('manager', 'access.grant.write', 'allow'),
  ('manager', 'access.simulate', 'allow'),
  ('manager', 'data_quality.read', 'allow'),
  ('analyst', 'data_quality.read', 'allow'),
  ('marketer', 'dashboard.read', 'allow'),
  ('manager', 'dashboard.read', 'allow'),
  ('analyst', 'dashboard.read', 'allow'),
  ('marketer', 'tenant.aggregates.read', 'allow'),
  ('manager', 'tenant.aggregates.read', 'allow'),
  ('analyst', 'tenant.aggregates.read', 'allow'),
  ('tenant', 'tenant.aggregates.own', 'allow'),
  ('marketer', 'metric.first_purchase_new_tenant.read', 'allow'),
  ('manager', 'metric.first_purchase_new_tenant.read', 'allow'),
  ('analyst', 'metric.first_purchase_new_tenant.read', 'allow'),
  ('marketer', 'metric.cross_tenant_distribution.read', 'allow'),
  ('manager', 'metric.cross_tenant_distribution.read', 'allow'),
  ('analyst', 'metric.cross_tenant_distribution.read', 'allow'),
  ('marketer', 'metric.cross_tenant_rate.read', 'allow'),
  ('manager', 'metric.cross_tenant_rate.read', 'allow'),
  ('analyst', 'metric.cross_tenant_rate.read', 'allow'),
  ('marketer', 'metric.match_rate.read', 'allow'),
  ('manager', 'metric.match_rate.read', 'allow'),
  ('analyst', 'metric.match_rate.read', 'allow'),
  ('marketer', 'metric.disputed_merge_rate.read', 'allow'),
  ('manager', 'metric.disputed_merge_rate.read', 'allow'),
  ('analyst', 'metric.disputed_merge_rate.read', 'allow'),
  ('marketer', 'metric.segment_to_export_time.read', 'allow'),
  ('manager', 'metric.segment_to_export_time.read', 'allow'),
  ('analyst', 'metric.segment_to_export_time.read', 'allow'),
  ('marketer', 'metric.conversion_rate.read', 'allow'),
  ('manager', 'metric.conversion_rate.read', 'allow'),
  ('analyst', 'metric.conversion_rate.read', 'allow'),
  ('marketer', 'metric.conversion_lift.read', 'allow'),
  ('manager', 'metric.conversion_lift.read', 'allow'),
  ('analyst', 'metric.conversion_lift.read', 'allow'),
  ('marketer', 'metric.new_guests_via_mall.read', 'allow'),
  ('manager', 'metric.new_guests_via_mall.read', 'allow'),
  ('tenant', 'metric.new_guests_via_mall.read', 'allow'),
  ('analyst', 'metric.new_guests_via_mall.read', 'allow'),
  ('marketer', 'metric.repeat_visits.read', 'allow'),
  ('manager', 'metric.repeat_visits.read', 'allow'),
  ('tenant', 'metric.repeat_visits.read', 'allow'),
  ('analyst', 'metric.repeat_visits.read', 'allow'),
  ('marketer', 'metric.revenue_total.read', 'allow'),
  ('manager', 'metric.revenue_total.read', 'allow'),
  ('analyst', 'metric.revenue_total.read', 'allow'),
  ('marketer', 'metric.avg_check.read', 'allow'),
  ('manager', 'metric.avg_check.read', 'allow'),
  ('analyst', 'metric.avg_check.read', 'allow'),
  ('marketer', 'metric.visits_count.read', 'allow'),
  ('manager', 'metric.visits_count.read', 'allow'),
  ('tenant', 'metric.visits_count.read', 'allow'),
  ('analyst', 'metric.visits_count.read', 'allow'),
  ('marketer', 'metric.segment_size.read', 'allow'),
  ('manager', 'metric.segment_size.read', 'allow'),
  ('analyst', 'metric.segment_size.read', 'allow'),
  ('marketer', 'metric.export_counts.read', 'allow'),
  ('manager', 'metric.export_counts.read', 'allow'),
  ('analyst', 'metric.export_counts.read', 'allow'),
  ('marketer', 'metric.contact_frequency.read', 'allow'),
  ('manager', 'metric.contact_frequency.read', 'allow'),
  ('analyst', 'metric.contact_frequency.read', 'allow'),
  ('manager', 'metric.data_quality.read', 'allow'),
  ('analyst', 'metric.data_quality.read', 'allow'),
  ('marketer', 'metric.merge_queue_size.read', 'allow'),
  ('manager', 'metric.merge_queue_size.read', 'allow'),
  ('analyst', 'metric.merge_queue_size.read', 'allow'),
  ('marketer', 'metric.campaign_results.read', 'allow'),
  ('manager', 'metric.campaign_results.read', 'allow'),
  ('marketer', 'metric.tenant_priority.read', 'allow'),
  ('manager', 'metric.tenant_priority.read', 'allow'),
  ('analyst', 'metric.tenant_priority.read', 'allow');

INSERT INTO role_permission (role, permission_code, effect)
SELECT roles.role, permission.code, 'allow'
FROM (SELECT 'marketer' AS role UNION ALL SELECT 'manager' UNION ALL SELECT 'analyst') roles
CROSS JOIN permission
WHERE permission.code LIKE 'field.guest.%' OR permission.code LIKE 'field.event.%';

INSERT INTO role_permission (role, permission_code, effect)
SELECT 'tenant', permission.code, 'allow'
FROM permission
WHERE permission.code IN ('guest.profile.read', 'guest.event.read', 'guest.identifier.read')
   OR permission.code LIKE 'field.guest.%' OR permission.code LIKE 'field.event.%';

INSERT INTO role_permission (role, permission_code, effect)
VALUES
  ('manager', 'guest.profile.read', 'allow'),
  ('manager', 'guest.profile.search', 'allow'),
  ('manager', 'guest.event.read', 'allow'),
  ('manager', 'guest.identifier.read', 'allow'),
  ('manager', 'guest.contact_history.read', 'allow'),
  ('marketer', 'field.export.contact.read', 'allow'),
  ('marketer', 'field.tenant.priority.read', 'allow'),
  ('manager', 'field.tenant.priority.read', 'allow'),
  ('analyst', 'field.tenant.priority.read', 'allow');

CREATE TABLE access_policy_state (
  id         INTEGER PRIMARY KEY CHECK (id = 1),
  version    BIGINT NOT NULL CHECK (version >= 0),
  updated_at TIMESTAMP NOT NULL
);
INSERT INTO access_policy_state (id, version, updated_at) VALUES (1, 0, CURRENT_TIMESTAMP);

CREATE TABLE access_grant (
  id             TEXT PRIMARY KEY,
  subject_type   VARCHAR(16) NOT NULL CHECK (subject_type IN ('user', 'tenant', 'role')),
  subject_id     TEXT NOT NULL,
  preset_code    VARCHAR(32),
  status         VARCHAR(16) NOT NULL CHECK (status IN ('draft', 'active', 'suspended', 'revoked', 'expired')),
  valid_from     TIMESTAMP NOT NULL,
  valid_to       TIMESTAMP,
  created_by     TEXT REFERENCES app_user(id),
  approved_by    TEXT REFERENCES app_user(id),
  reason         TEXT,
  policy_version BIGINT NOT NULL,
  created_at     TIMESTAMP NOT NULL,
  updated_at     TIMESTAMP NOT NULL,
  CHECK (valid_to IS NULL OR valid_to > valid_from)
);
CREATE INDEX ix_grant_subject ON access_grant(subject_type, subject_id, status);
CREATE INDEX ix_grant_window ON access_grant(status, valid_from, valid_to);

CREATE TABLE access_grant_item (
  grant_id        TEXT NOT NULL REFERENCES access_grant(id),
  permission_code VARCHAR(64) NOT NULL REFERENCES permission(code),
  effect          VARCHAR(8) NOT NULL CHECK (effect IN ('allow', 'deny')),
  PRIMARY KEY (grant_id, permission_code)
);

CREATE TABLE access_grant_scope (
  grant_id    TEXT NOT NULL REFERENCES access_grant(id),
  scope_type  VARCHAR(24) NOT NULL CHECK (scope_type IN
                ('tenant', 'category', 'period_days', 'channel', 'k_threshold', 'row_limit', 'field_mask', 'metric', 'principal_role')),
  scope_value TEXT NOT NULL,
  PRIMARY KEY (grant_id, scope_type, scope_value)
);

UPDATE access_policy_state SET version = 1, updated_at = CURRENT_TIMESTAMP WHERE id = 1;

INSERT INTO access_grant (id, subject_type, subject_id, status, valid_from, reason, policy_version, created_at, updated_at)
SELECT 'bootstrap-role-' || role, 'role', role, 'active', CURRENT_TIMESTAMP,
       'Initial role access migration', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM (SELECT 'marketer' AS role UNION ALL SELECT 'manager' UNION ALL SELECT 'analyst') roles;

INSERT INTO access_grant_item (grant_id, permission_code, effect)
SELECT 'bootstrap-role-' || role, permission_code, effect
FROM role_permission
WHERE role IN ('marketer', 'manager', 'analyst')
  AND (role <> 'manager' OR (
    permission_code NOT LIKE 'field.guest.%' AND permission_code NOT LIKE 'field.event.%'
    AND permission_code NOT IN ('guest.profile.read', 'guest.profile.search', 'guest.event.read',
                                'guest.identifier.read', 'guest.contact_history.read')
  ));

INSERT INTO access_grant (id, subject_type, subject_id, status, valid_from, reason, policy_version, created_at, updated_at)
SELECT 'bootstrap-manager-' || md5(id), 'user', id, 'active', CURRENT_TIMESTAMP,
       'Initial personal data access migration', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM app_user WHERE role = 'manager' AND is_active;

INSERT INTO access_grant_item (grant_id, permission_code, effect)
SELECT g.id, rp.permission_code, rp.effect
FROM access_grant g
JOIN role_permission rp ON rp.role = 'manager'
WHERE g.subject_type = 'user' AND g.id LIKE 'bootstrap-manager-%'
  AND (rp.permission_code LIKE 'field.guest.%' OR rp.permission_code LIKE 'field.event.%'
    OR rp.permission_code IN ('guest.profile.read', 'guest.profile.search', 'guest.event.read',
                              'guest.identifier.read', 'guest.contact_history.read'));

INSERT INTO access_grant_scope (grant_id, scope_type, scope_value)
SELECT id, 'principal_role', 'manager'
FROM access_grant WHERE subject_type = 'user' AND id LIKE 'bootstrap-manager-%';

INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, created_at)
SELECT 'bootstrap-access-audit-' || md5(id), NULL, 'system', 'access.grant.create',
       'access_grant', id, '{"origin":"migration","policyVersion":1}', CURRENT_TIMESTAMP
FROM access_grant WHERE id LIKE 'bootstrap-role-%' OR id LIKE 'bootstrap-manager-%';

-- +goose Down
DELETE FROM audit_log WHERE actor_role = 'system' AND action = 'access.grant.create'
  AND id LIKE 'bootstrap-access-audit-%';
DROP TABLE IF EXISTS access_grant_scope;
DROP TABLE IF EXISTS access_grant_item;
DROP TABLE IF EXISTS access_grant;
DROP TABLE IF EXISTS access_policy_state;
DROP TABLE IF EXISTS role_permission;
DROP TABLE IF EXISTS permission;
