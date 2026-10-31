package ingestion

import (
	"encoding/json"
	"strings"
)

type FieldMapping struct {
	ExternalID  string `json:"externalId"`
	OccurredAt  string `json:"occurredAt"`
	EventType   string `json:"eventType"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Card        string `json:"card"`
	Name        string `json:"name"`
	TenantCode  string `json:"tenantCode"`
	ProjectCode string `json:"projectCode"`
}

func DefaultFieldMapping() FieldMapping {
	return FieldMapping{
		ExternalID:  "external_id",
		OccurredAt:  "occurred_at",
		EventType:   "event_type",
		Amount:      "amount",
		Currency:    "currency",
		Phone:       "phone",
		Email:       "email",
		Card:        "card",
		Name:        "name",
		TenantCode:  "tenant_code",
		ProjectCode: "project_code",
	}
}

func ParseFieldMapping(raw string) (FieldMapping, error) {
	m := DefaultFieldMapping()
	if strings.TrimSpace(raw) == "" {
		return m, nil
	}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return FieldMapping{}, err
	}
	def := DefaultFieldMapping()
	if m.ExternalID == "" {
		m.ExternalID = def.ExternalID
	}
	if m.OccurredAt == "" {
		m.OccurredAt = def.OccurredAt
	}
	if m.EventType == "" {
		m.EventType = def.EventType
	}
	return m, nil
}

func (m FieldMapping) requiredColumns() []string {
	return []string{m.ExternalID, m.OccurredAt, m.EventType}
}
