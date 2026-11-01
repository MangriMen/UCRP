package access

import (
	"sort"
	"strconv"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type Resource struct {
	TenantID   string
	Category   string
	Channel    string
	Metric     string
	PeriodDays int
	RowCount   int
}

type Decision struct {
	Allowed    bool
	Reason     string
	PeriodDays int
	KThreshold int
	RowLimit   int
	MaskFields []string
}

func Decide(principal domain.Principal, permission string, resource Resource, ceiling []domain.AccessRolePermission, grants []domain.AccessGrant, now time.Time) Decision {
	if !InCeiling(permission, ceiling) {
		return Decision{Reason: "forbidden_permission"}
	}
	if principal.Role == domain.RoleTenant && (principal.TenantID == "" || principal.TenantID != resource.TenantID) {
		return Decision{Reason: "forbidden_tenant"}
	}

	decision := Decision{Reason: "forbidden_permission", KThreshold: 5}
	found := false
	for _, grant := range grants {
		if !activeFor(grant, principal, now) {
			continue
		}
		for _, item := range grant.Items {
			if item.PermissionCode != permission {
				continue
			}
			found = true
			matched, valid, limits := matchScope(grant.Scopes, principal, resource)
			if !valid {
				return Decision{Reason: "forbidden_scope"}
			}
			if !matched {
				continue
			}
			switch item.Effect {
			case "deny":
				return Decision{Reason: "forbidden_scope"}
			case "allow":
				if principal.Role == domain.RoleTenant && !hasTenantScope(grant.Scopes) {
					continue
				}
				decision.Allowed = true
				decision.PeriodDays = minimumLimit(decision.PeriodDays, limits.PeriodDays)
				decision.RowLimit = minimumLimit(decision.RowLimit, limits.RowLimit)
				if limits.KThreshold > decision.KThreshold {
					decision.KThreshold = limits.KThreshold
				}
				decision.MaskFields = append(decision.MaskFields, limits.MaskFields...)
			default:
				return Decision{Reason: "forbidden_scope"}
			}
		}
	}
	if !decision.Allowed {
		if found {
			return Decision{Reason: "forbidden_scope"}
		}
		return Decision{Reason: "forbidden_permission"}
	}
	if !withinLimits(resource, decision) {
		return Decision{Reason: "forbidden_scope"}
	}
	decision.Reason = ""
	sort.Strings(decision.MaskFields)
	decision.MaskFields = uniqueStrings(decision.MaskFields)
	return decision
}

func hasTenantScope(scopes []domain.AccessGrantScope) bool {
	for _, scope := range scopes {
		if scope.Type == "tenant" {
			return true
		}
	}
	return false
}

func InCeiling(permission string, ceiling []domain.AccessRolePermission) bool {
	allowed := false
	for _, entry := range ceiling {
		if entry.PermissionCode != permission {
			continue
		}
		if entry.Effect != "allow" {
			return false
		}
		allowed = true
	}
	return allowed
}

func activeFor(grant domain.AccessGrant, principal domain.Principal, now time.Time) bool {
	if grant.Status != "active" || now.Before(grant.ValidFrom) || (grant.ValidTo != nil && !now.Before(*grant.ValidTo)) {
		return false
	}
	switch grant.SubjectType {
	case "user":
		return principal.UserID != "" && grant.SubjectID == principal.UserID
	case "tenant":
		return principal.TenantID != "" && grant.SubjectID == principal.TenantID
	case "role":
		return principal.Role != "" && grant.SubjectID == string(principal.Role)
	default:
		return false
	}
}

func matchScope(scopes []domain.AccessGrantScope, principal domain.Principal, resource Resource) (bool, bool, Decision) {
	limits := Decision{KThreshold: 5}
	present := make(map[string]bool)
	matched := make(map[string]bool)
	for _, scope := range scopes {
		switch scope.Type {
		case "tenant", "category", "channel", "metric", "principal_role":
			if scope.Value == "" {
				return false, false, Decision{}
			}
			present[scope.Type] = true
			value := scope.Value
			if scope.Type == "tenant" && value == "self" {
				value = principal.TenantID
				if value == "" {
					return false, false, Decision{}
				}
			}
			switch scope.Type {
			case "tenant":
				matched[scope.Type] = matched[scope.Type] || value == resource.TenantID
			case "category":
				matched[scope.Type] = matched[scope.Type] || value == resource.Category
			case "metric":
				matched[scope.Type] = matched[scope.Type] || value == resource.Metric
			case "principal_role":
				matched[scope.Type] = matched[scope.Type] || value == string(principal.Role)
			case "channel":
				if value != "sms" && value != "push" && value != "email" {
					return false, false, Decision{}
				}
				matched[scope.Type] = matched[scope.Type] || value == resource.Channel
			}
		case "period_days", "row_limit", "k_threshold":
			value, err := strconv.Atoi(scope.Value)
			if err != nil || value <= 0 || (scope.Type == "k_threshold" && value < 5) {
				return false, false, Decision{}
			}
			switch scope.Type {
			case "period_days":
				limits.PeriodDays = minimumLimit(limits.PeriodDays, value)
			case "row_limit":
				limits.RowLimit = minimumLimit(limits.RowLimit, value)
			case "k_threshold":
				if value > limits.KThreshold {
					limits.KThreshold = value
				}
			}
		case "field_mask":
			if scope.Value != "phone" && scope.Value != "email" && scope.Value != "amount" {
				return false, false, Decision{}
			}
			limits.MaskFields = append(limits.MaskFields, scope.Value)
		default:
			return false, false, Decision{}
		}
	}
	for kind := range present {
		if !matched[kind] {
			return false, true, limits
		}
	}
	return true, true, limits
}

func withinLimits(resource Resource, limits Decision) bool {
	return resource.PeriodDays >= 0 && resource.RowCount >= 0 &&
		(limits.PeriodDays == 0 || resource.PeriodDays <= limits.PeriodDays) &&
		(limits.RowLimit == 0 || resource.RowCount <= limits.RowLimit)
}

func minimumLimit(current, value int) int {
	if current == 0 || (value > 0 && value < current) {
		return value
	}
	return current
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}
