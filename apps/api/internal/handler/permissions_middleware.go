package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
)

type protectedRoute struct {
	method      string
	path        string
	permissions []string
}

type policyAuthorizer interface {
	accessAuthorizer
	PolicyVersion(context.Context) (int64, error)
}

func PermissionMiddleware(authorizer policyAuthorizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := auth.PrincipalFrom(r.Context())
			if !ok {
				httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
				return
			}
			path := strings.TrimPrefix(r.URL.Path, "/api/v1")
			if path == "/auth/me" || path == "/access" || strings.HasPrefix(path, "/access/") {
				next.ServeHTTP(w, r)
				return
			}
			parts := pathParts(path)
			if principal.Role == domain.RoleTenant && r.Method == http.MethodGet && len(parts) >= 1 && len(parts) <= 2 && parts[0] == "tenants" {
				next.ServeHTTP(w, r)
				return
			}
			if principal.Role == domain.RoleTenant && r.Method == http.MethodGet && len(parts) == 3 && parts[0] == "tenants" && parts[2] == "aggregates" {
				next.ServeHTTP(w, r)
				return
			}
			if principal.Role == domain.RoleTenant && r.Method == http.MethodGet && len(parts) >= 2 && len(parts) <= 3 && parts[0] == "guests" && (len(parts) == 2 || parts[2] == "events") {
				next.ServeHTTP(w, r)
				return
			}
			permissions, ok := routePermissions(r.Method, parts)
			if !ok || authorizer == nil {
				httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа")
				return
			}
			for _, permission := range permissions {
				segmentCount := len(parts) > 0 && parts[0] == "segments"
				exportCount := len(parts) > 0 && len(parts) <= 2 && parts[0] == "exports"
				dashboard := r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "dashboard" && (parts[1] == "cross-tenant" || parts[1] == "metrics")
				dataQuality := r.Method == http.MethodGet && len(parts) >= 2 && len(parts) <= 3 && parts[0] == "data-quality" && parts[1] == "sources"
				if strings.HasPrefix(permission, "metric.") && !segmentCount && !exportCount && !dashboard && !dataQuality && (len(parts) != 3 || parts[0] != "tenants" || parts[2] != "aggregates") {
					httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа")
					return
				}
			}
			version, err := authorizer.PolicyVersion(r.Context())
			if err != nil {
				httpx.WriteDomainError(w, r, err)
				return
			}
			resource := access.Resource{}
			if len(parts) == 3 && parts[0] == "tenants" && parts[2] == "aggregates" {
				resource.TenantID = parts[1]
			}
			for _, permission := range permissions {
				resource.Metric = ""
				if strings.HasPrefix(permission, "metric.") {
					resource.Metric = permission
				}
				decision, err := authorizer.Authorize(r.Context(), principal, permission, resource)
				if err != nil {
					httpx.WriteDomainError(w, r, err)
					return
				}
				if !decision.Allowed || decision.PeriodDays > 0 || decision.RowLimit > 0 || decision.KThreshold > 5 || len(decision.MaskFields) > 0 {
					httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа")
					return
				}
			}
			currentVersion, err := authorizer.PolicyVersion(r.Context())
			if err != nil {
				httpx.WriteDomainError(w, r, err)
				return
			}
			if version != currentVersion {
				httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func pathParts(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func routePermissions(method string, parts []string) ([]string, bool) {
	for _, route := range protectedRoutes {
		if route.method != method {
			continue
		}
		pattern := pathParts(route.path)
		if len(parts) != len(pattern) {
			continue
		}
		matched := true
		for i := range pattern {
			if pattern[i] != parts[i] && pattern[i] != "{id}" {
				matched = false
				break
			}
		}
		if matched {
			return route.permissions, true
		}
	}
	return nil, false
}

var protectedRoutes = []protectedRoute{
	{http.MethodGet, "/guests", []string{"guest.profile.search", "guest.identifier.read", "field.guest.display_name.read", "field.guest.tenants_visited.read", "field.guest.phone.read", "field.guest.email.read", "field.guest.card.read"}},
	{http.MethodGet, "/guests/{id}/events", []string{"guest.event.read", "field.event.amount.read", "field.event.external_id.read"}},
	{http.MethodGet, "/guests/{id}/movement", []string{"guest.event.read"}},
	{http.MethodGet, "/guests/{id}/contacts", []string{"guest.contact_history.read"}},
	{http.MethodPost, "/guests/{id}/merge", []string{"identity.decide"}},
	{http.MethodDelete, "/guests/{id}", []string{"guest.profile.delete"}},
	{http.MethodGet, "/guests/{id}/recommendations", []string{"recommendation.read"}},
	{http.MethodPost, "/guests/{id}/recommendations/refresh", []string{"recommendation.refresh"}},
	{http.MethodGet, "/guests/{id}", []string{"guest.profile.read", "guest.identifier.read", "field.guest.display_name.read", "field.guest.tenants_visited.read", "field.guest.identifiers.read", "field.guest.phone.read", "field.guest.email.read", "field.guest.card.read"}},
	{http.MethodGet, "/merge-candidates", []string{"identity.queue.read"}},
	{http.MethodGet, "/merge-candidates/{id}", []string{"identity.queue.read"}},
	{http.MethodPost, "/merge-candidates/{id}/confirm", []string{"identity.decide"}},
	{http.MethodPost, "/merge-candidates/{id}/reject", []string{"identity.decide"}},
	{http.MethodPost, "/merge-candidates/{id}/revert", []string{"identity.decide"}},
	{http.MethodGet, "/merge-blocks", []string{"identity.queue.read"}},
	{http.MethodDelete, "/merge-blocks/{id}", []string{"identity.decide"}},
	{http.MethodPost, "/segments/preview", []string{"segment.preview", "metric.segment_size.read"}},
	{http.MethodGet, "/segments", []string{"segment.read", "metric.segment_size.read"}},
	{http.MethodPost, "/segments", []string{"segment.write", "metric.segment_size.read"}},
	{http.MethodGet, "/segments/{id}/members", []string{"segment.members.read"}},
	{http.MethodGet, "/segments/{id}/versions", []string{"segment.read", "metric.segment_size.read"}},
	{http.MethodPost, "/segments/{id}/recompute", []string{"segment.write", "metric.segment_size.read"}},
	{http.MethodGet, "/segments/{id}/recommendations", []string{"recommendation.read", "metric.segment_size.read"}},
	{http.MethodGet, "/segments/{id}", []string{"segment.read", "metric.segment_size.read"}},
	{http.MethodPut, "/segments/{id}", []string{"segment.write", "metric.segment_size.read"}},
	{http.MethodDelete, "/segments/{id}", []string{"segment.delete"}},
	{http.MethodGet, "/exports", []string{"export.read", "metric.export_counts.read"}},
	{http.MethodPost, "/exports/preview", []string{"export.create", "metric.export_counts.read"}},
	{http.MethodPost, "/exports", []string{"export.create", "metric.export_counts.read"}},
	{http.MethodGet, "/exports/{id}/members", []string{"export.members.read", "field.export.contact.read"}},
	{http.MethodGet, "/exports/{id}/download", []string{"export.download", "field.export.contact.read"}},
	{http.MethodGet, "/exports/{id}", []string{"export.read", "metric.export_counts.read"}},
	{http.MethodGet, "/contact-limits", []string{"contact_limit.read"}},
	{http.MethodPost, "/contact-limits", []string{"contact_limit.write"}},
	{http.MethodPatch, "/contact-limits/{id}", []string{"contact_limit.write"}},
	{http.MethodDelete, "/contact-limits/{id}", []string{"contact_limit.write"}},
	{http.MethodGet, "/tenants/{id}/guests", []string{"guest.profile.search", "guest.identifier.read", "field.guest.display_name.read", "field.guest.tenants_visited.read", "field.guest.phone.read", "field.guest.email.read", "field.guest.card.read"}},
	{http.MethodGet, "/tenants/{id}/aggregates", []string{"tenant.aggregates.read", "metric.visits_count.read"}},
	{http.MethodGet, "/tenants", []string{"reference.read", "field.tenant.priority.read"}},
	{http.MethodGet, "/tenants/{id}", []string{"reference.read", "field.tenant.priority.read"}},
	{http.MethodPost, "/tenants", []string{"reference.write"}},
	{http.MethodPatch, "/tenants/{id}", []string{"reference.write"}},
	{http.MethodDelete, "/tenants/{id}", []string{"reference.write"}},
	{http.MethodGet, "/malls", []string{"reference.read"}},
	{http.MethodPost, "/malls", []string{"reference.write"}},
	{http.MethodPatch, "/malls/{id}", []string{"reference.write"}},
	{http.MethodDelete, "/malls/{id}", []string{"reference.write"}},
	{http.MethodGet, "/tenant-categories", []string{"reference.read"}},
	{http.MethodPost, "/tenant-categories", []string{"reference.write"}},
	{http.MethodPatch, "/tenant-categories/{id}", []string{"reference.write"}},
	{http.MethodDelete, "/tenant-categories/{id}", []string{"reference.write"}},
	{http.MethodGet, "/tenant-categories/{id}/relations", []string{"reference.read"}},
	{http.MethodPut, "/tenant-categories/{id}/relations", []string{"reference.write"}},
	{http.MethodGet, "/projects", []string{"reference.read"}},
	{http.MethodPost, "/projects", []string{"reference.write"}},
	{http.MethodPatch, "/projects/{id}", []string{"reference.write"}},
	{http.MethodDelete, "/projects/{id}", []string{"reference.write"}},
	{http.MethodGet, "/event-types", []string{"reference.read"}},
	{http.MethodPost, "/event-types", []string{"reference.write"}},
	{http.MethodPatch, "/event-types/{id}", []string{"reference.write"}},
	{http.MethodDelete, "/event-types/{id}", []string{"reference.write"}},
	{http.MethodPost, "/imports", []string{"import.create"}},
	{http.MethodGet, "/imports", []string{"import.read"}},
	{http.MethodGet, "/imports/{id}/rejected-rows", []string{"import.rejected_rows.read", "field.guest.phone.read", "field.guest.email.read", "field.guest.card.read", "field.event.raw_payload.read"}},
	{http.MethodGet, "/imports/{id}", []string{"import.read"}},
	{http.MethodPost, "/data-sources/{id}/sync", []string{"import.create"}},
	{http.MethodGet, "/data-sources", []string{"data_source.read"}},
	{http.MethodPost, "/data-sources", []string{"data_source.write"}},
	{http.MethodGet, "/data-sources/{id}", []string{"data_source.read"}},
	{http.MethodPatch, "/data-sources/{id}", []string{"data_source.write"}},
	{http.MethodDelete, "/data-sources/{id}", []string{"data_source.write"}},
	{http.MethodGet, "/segment-templates", []string{"segment.read"}},
	{http.MethodGet, "/data-source-types", []string{"data_source.read"}},
	{http.MethodPost, "/data-source-types", []string{"data_source.write"}},
	{http.MethodPatch, "/data-source-types/{id}", []string{"data_source.write"}},
	{http.MethodGet, "/users", []string{"user.read"}},
	{http.MethodPost, "/users", []string{"user.write"}},
	{http.MethodPost, "/users/invitations", []string{"user.write"}},
	{http.MethodPost, "/users/{id}/recovery", []string{"user.write"}},
	{http.MethodGet, "/users/{id}", []string{"user.read"}},
	{http.MethodPatch, "/users/{id}", []string{"user.write"}},
	{http.MethodDelete, "/users/{id}", []string{"user.write"}},
	{http.MethodGet, "/audit-logs", []string{"audit.read", "field.guest.phone.read", "field.guest.email.read", "field.guest.card.read", "field.event.raw_payload.read"}},
	{http.MethodGet, "/campaigns", []string{"campaign.read"}},
	{http.MethodPost, "/campaigns", []string{"campaign.write"}},
	{http.MethodGet, "/campaigns/{id}", []string{"campaign.read"}},
	{http.MethodGet, "/campaigns/{id}/results", []string{"campaign.read"}},
	{http.MethodPatch, "/campaigns/{id}", []string{"campaign.write"}},
	{http.MethodDelete, "/campaigns/{id}", []string{"campaign.write"}},
	{http.MethodGet, "/dashboard/cross-tenant", []string{"dashboard.read", "metric.cross_tenant_distribution.read", "metric.cross_tenant_rate.read"}},
	{http.MethodGet, "/dashboard/metrics", []string{"dashboard.read", "metric.match_rate.read", "metric.disputed_merge_rate.read", "metric.segment_to_export_time.read"}},
	{http.MethodGet, "/data-quality/sources", []string{"data_quality.read", "metric.data_quality.read", "metric.merge_queue_size.read"}},
	{http.MethodGet, "/data-quality/sources/{id}", []string{"data_quality.read", "metric.data_quality.read", "metric.merge_queue_size.read"}},
}
