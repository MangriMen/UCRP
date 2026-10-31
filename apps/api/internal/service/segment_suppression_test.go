package service

import (
	"context"
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

type segmentPreviewEvaluator struct {
	size int
}

func (e segmentPreviewEvaluator) Preview(context.Context, domain.SegmentRule) (int, []string, error) {
	return e.size, []string{"guest-1", "guest-2"}, nil
}

func (e segmentPreviewEvaluator) Evaluate(context.Context, domain.SegmentRule) ([]string, error) {
	return nil, nil
}

type segmentVersionRepo struct {
	fakeSegments
	versions []domain.SegmentVersion
}

func (r *segmentVersionRepo) ListVersions(context.Context, string) ([]domain.SegmentVersion, error) {
	return r.versions, nil
}

func TestSegmentCountSuppression(t *testing.T) {
	four, five := 4, 5
	for _, tc := range []struct {
		name       string
		count      *int
		suppressed bool
	}{
		{"unknown", nil, true},
		{"four", &four, true},
		{"five", &five, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dto := toSegmentDTO(&domain.Segment{}, &domain.SegmentVersion{MemberCount: tc.count})
			if dto.Suppressed != tc.suppressed || (dto.MemberCount == nil) != tc.suppressed {
				t.Fatalf("unexpected segment: %+v", dto)
			}
			if !tc.suppressed && *dto.MemberCount != *tc.count {
				t.Fatalf("unexpected count: %d", *dto.MemberCount)
			}
		})
	}
}

func TestSegmentPreviewSuppression(t *testing.T) {
	for _, size := range []int{0, 4, 5} {
		service := NewSegmentService(nil, segmentPreviewEvaluator{size: size}, nil)
		response, err := service.Preview(context.Background(), SegmentPreviewRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if size < 5 {
			if response.Size != nil || !response.Suppressed || len(response.Examples) != 0 {
				t.Fatalf("size %d leaked: %+v", size, response)
			}
		} else if response.Size == nil || *response.Size != size || response.Suppressed || len(response.Examples) != 2 {
			t.Fatalf("size %d incorrectly suppressed: %+v", size, response)
		}
	}
}

func TestSegmentVersionSuppression(t *testing.T) {
	four, five := 4, 5
	repo := &segmentVersionRepo{versions: []domain.SegmentVersion{{}, {MemberCount: &four}, {MemberCount: &five}}}
	service := NewSegmentService(repo, nil, nil)
	versions, err := service.Versions(context.Background(), "segment")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 || !versions[0].Suppressed || versions[0].MemberCount != nil || !versions[1].Suppressed || versions[1].MemberCount != nil || versions[2].Suppressed || versions[2].MemberCount == nil || *versions[2].MemberCount != 5 {
		t.Fatalf("unexpected versions: %+v", versions)
	}
}
