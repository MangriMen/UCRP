package seed

import (
	"context"
	"encoding/json"
)

type jobSpec struct {
	Key      string
	JobType  string
	Payload  map[string]string
	DedupKey string
	DaysAgo  int
}

func demoJobs() []jobSpec {
	return []jobSpec{
		{"job-import", "import.process", map[string]string{"batch_id": "bean-1030"}, "import:bean-1030", 1},
		{"job-identity", "identity.resolve", map[string]string{"batch_id": "cinema-1020"}, "identity:cinema-1020", 0},
	}
}

func (s *Seeder) seedJobs(ctx context.Context) error {
	for _, j := range demoJobs() {
		payload := map[string]string{}
		for k, v := range j.Payload {
			if k == "batch_id" {
				payload[k] = s.batches[v]
				continue
			}
			payload[k] = v
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if err := s.exec(ctx, `
			INSERT INTO job_queue (id, job_type, payload, dedup_key, status, attempts, run_after, created_at)
			VALUES ($1,$2,$3,$4,'queued',0,$5,$5)
			ON CONFLICT (id) DO NOTHING`,
			demoID("job_queue", j.Key), j.JobType, string(raw), j.DedupKey,
			s.now.AddDate(0, 0, -j.DaysAgo)); err != nil {
			return err
		}
	}
	return nil
}
