package service

import "testing"

func TestStoredRuleJSON_DropsNullListsOfOldVersions(t *testing.T) {
	cases := map[string]string{
		`{"visitedTenantIds":["a"],"excludedTenantIds":null,"periodDays":30}`:                                 `{"visitedTenantIds":["a"],"periodDays":30}`,
		`{"operator":"or","conditions":[{"visitedTenantIds":["a"],"excludedTenantIds":null,"periodDays":7}]}`: `{"operator":"or","conditions":[{"visitedTenantIds":["a"],"periodDays":7}]}`,
		`{"visitedTenantIds":["a"],"excluded_tenant_ids":["b"],"period_days":14}`:                             `{"visitedTenantIds":["a"],"excludedTenantIds":["b"],"periodDays":14}`,
		`not json`: `not json`,
	}
	for stored, want := range cases {
		if got := string(storedRuleJSON(stored)); got != want {
			t.Errorf("storedRuleJSON(%s) = %s, want %s", stored, got, want)
		}
	}
}
