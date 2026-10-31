package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

type stubRelationRepo struct {
	stubCategoryRepo
	current  []domain.CategoryRelation
	replaced []domain.CategoryRelation
	called   bool
}

func (r *stubRelationRepo) ListRelations(context.Context, string) ([]domain.CategoryRelation, error) {
	if r.called {
		return r.replaced, nil
	}
	return r.current, nil
}

func (r *stubRelationRepo) ReplaceRelations(_ context.Context, _ string, rels []domain.CategoryRelation) error {
	r.called = true
	r.replaced = rels
	return nil
}

func newRelationService(current ...domain.CategoryRelation) (*ReferenceService, *stubRelationRepo) {
	repo := &stubRelationRepo{
		stubCategoryRepo: stubCategoryRepo{categories: map[string]domain.TenantCategory{
			"food":     {ID: "food", IsActive: true},
			"cinema":   {ID: "cinema", IsActive: true},
			"fashion":  {ID: "fashion", IsActive: true},
			"inactive": {ID: "inactive", IsActive: false},
		}},
		current: current,
	}
	return NewReferenceService(nil, nil, repo, nil, nil), repo
}

func relation(related, relationType string, weight *int) RelationDTO {
	return RelationDTO{CategoryID: "food", RelatedCategoryID: related, RelationType: relationType, WeightPct: weight}
}

func TestReplaceRelations_RejectsInvalidSet(t *testing.T) {
	overRange := 101
	negative := -1
	tests := []struct {
		name string
		rels []RelationDTO
	}{
		{"unknown type", []RelationDTO{relation("cinema", "bogus", nil)}},
		{"self relation", []RelationDTO{relation("food", "adjacent", nil)}},
		{"foreign category id", []RelationDTO{{CategoryID: "cinema", RelatedCategoryID: "fashion", RelationType: "adjacent"}}},
		{"empty related id", []RelationDTO{relation(" ", "adjacent", nil)}},
		{"unknown related category", []RelationDTO{relation("missing", "adjacent", nil)}},
		{"new inactive related category", []RelationDTO{relation("inactive", "adjacent", nil)}},
		{"duplicate pair", []RelationDTO{relation("cinema", "adjacent", nil), relation("cinema", "adjacent", nil)}},
		{"pair with both types", []RelationDTO{relation("cinema", "adjacent", nil), relation("cinema", "competitor", nil)}},
		{"weight above range", []RelationDTO{relation("cinema", "adjacent", &overRange)}},
		{"negative weight", []RelationDTO{relation("cinema", "adjacent", &negative)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newRelationService()
			_, err := svc.ReplaceRelations(context.Background(), "food", tt.rels)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("err = %v, want validation", err)
			}
			if repo.called {
				t.Fatal("relations replaced despite invalid input")
			}
		})
	}
}

func TestReplaceRelations_ReturnsNotFound_WhenCategoryIsUnknown(t *testing.T) {
	svc, _ := newRelationService()
	if _, err := svc.ReplaceRelations(context.Background(), "missing", nil); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestReplaceRelations_DefaultsWeightAndKeepsZero(t *testing.T) {
	zero := 0
	svc, repo := newRelationService()
	_, err := svc.ReplaceRelations(context.Background(), "food", []RelationDTO{
		relation("cinema", "adjacent", nil),
		relation("fashion", "competitor", &zero),
	})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}
	if got := repo.replaced[0].WeightPct; got != defaultRelationWeightPct {
		t.Fatalf("omitted weight = %d, want %d", got, defaultRelationWeightPct)
	}
	if got := repo.replaced[1].WeightPct; got != 0 {
		t.Fatalf("zero weight = %d, want 0", got)
	}
}

func TestReplaceRelations_KeepsExistingRelationToInactiveCategory(t *testing.T) {
	svc, repo := newRelationService(domain.CategoryRelation{CategoryID: "food", RelatedCategoryID: "inactive", RelationType: "adjacent", WeightPct: 40})
	weight := 40
	if _, err := svc.ReplaceRelations(context.Background(), "food", []RelationDTO{relation("inactive", "adjacent", &weight)}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if len(repo.replaced) != 1 || repo.replaced[0].RelatedCategoryID != "inactive" {
		t.Fatalf("replaced = %+v", repo.replaced)
	}
}

func TestReplaceRelations_ClearsSet_WhenListIsEmpty(t *testing.T) {
	svc, repo := newRelationService(domain.CategoryRelation{CategoryID: "food", RelatedCategoryID: "cinema", RelationType: "adjacent", WeightPct: 100})
	out, err := svc.ReplaceRelations(context.Background(), "food", []RelationDTO{})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}
	if !repo.called || len(repo.replaced) != 0 || len(out) != 0 {
		t.Fatalf("called = %v, replaced = %+v, out = %+v", repo.called, repo.replaced, out)
	}
}
