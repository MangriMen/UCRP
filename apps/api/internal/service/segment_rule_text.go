package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/MangriMen/UCRP/internal/domain"
)

func parseRule(raw string) (domain.SegmentRule, error) {
	var rule domain.SegmentRule
	if err := json.Unmarshal([]byte(raw), &rule); err != nil {
		return domain.SegmentRule{}, err
	}
	return rule, nil
}

func (s *SegmentService) ruleTexts(ctx context.Context, rawRules []string) ([]string, error) {
	rules := make([]domain.SegmentRule, 0, len(rawRules))
	ids := map[string]bool{}
	for _, raw := range rawRules {
		if raw == "" {
			rules = append(rules, domain.SegmentRule{})
			continue
		}
		rule, err := parseRule(raw)
		if err != nil {
			rules = append(rules, domain.SegmentRule{})
			continue
		}
		rules = append(rules, rule)
		collectRuleTenantIDs(rule, ids)
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	names, err := s.tenants.Names(ctx, idList)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rules))
	for _, rule := range rules {
		out = append(out, ruleTextOf(rule, names))
	}
	return out, nil
}

func collectRuleTenantIDs(rule domain.SegmentRule, into map[string]bool) {
	if rule.IsGroup() {
		for _, cond := range rule.Conditions {
			collectRuleTenantIDs(cond, into)
		}
		return
	}
	for _, id := range rule.VisitedTenantIDs {
		into[id] = true
	}
	for _, id := range rule.ExcludedTenantIDs {
		into[id] = true
	}
}

func ruleTextOf(rule domain.SegmentRule, names map[string]string) string {
	if rule.IsGroup() {
		return groupRuleText(rule, names)
	}
	parts := make([]string, 0, 5)
	if len(rule.VisitedTenantIDs) > 0 {
		parts = append(parts, "Был в "+quotedNames(rule.VisitedTenantIDs, names))
	}
	if len(rule.ExcludedTenantIDs) > 0 {
		parts = append(parts, "не был в "+quotedNames(rule.ExcludedTenantIDs, names))
	}
	if floors := zoneFloors(rule.PurchasedZones); len(floors) > 0 {
		parts = append(parts, zonePhrase(floors))
	}
	switch rule.RequireContact {
	case "phone":
		parts = append(parts, "контакт: телефон")
	case "email":
		parts = append(parts, "контакт: email")
	}
	if rule.PeriodDays > 0 {
		parts = append(parts, fmt.Sprintf("за %d дн.", rule.PeriodDays))
	}
	return strings.Join(parts, ", ")
}

func groupRuleText(rule domain.SegmentRule, names map[string]string) string {
	sep := " и "
	if rule.Operator == "or" {
		sep = " или "
	}
	parts := make([]string, 0, len(rule.Conditions))
	for _, cond := range rule.Conditions {
		text := ruleTextOf(cond, names)
		if text == "" {
			continue
		}
		parts = append(parts, "("+text+")")
	}
	return strings.Join(parts, sep)
}

func quotedNames(ids []string, names map[string]string) string {
	quoted := make([]string, 0, len(ids))
	for _, id := range ids {
		name := id
		if n, ok := names[id]; ok && n != "" {
			name = n
		}
		quoted = append(quoted, "«"+name+"»")
	}
	return strings.Join(quoted, ", ")
}

func zoneFloors(zones []domain.ZoneRef) []int {
	seen := map[int]bool{}
	floors := make([]int, 0, len(zones))
	for _, z := range zones {
		if !seen[z.Floor] {
			seen[z.Floor] = true
			floors = append(floors, z.Floor)
		}
	}
	sort.Ints(floors)
	return floors
}

func zonePhrase(floors []int) string {
	if len(floors) == 1 {
		return fmt.Sprintf("покупал на этаже %d", floors[0])
	}
	nums := make([]string, 0, len(floors))
	for _, f := range floors {
		nums = append(nums, fmt.Sprintf("%d", f))
	}
	return "покупал на этажах " + strings.Join(nums, ", ")
}
