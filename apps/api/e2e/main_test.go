package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/MangriMen/UCRP/e2e/report"
)

func TestMain(m *testing.M) {
	code := m.Run()

	if os.Getenv("UCRP_E2E_BASE_URL") == "" {
		os.Exit(code)
	}

	path := os.Getenv("UCRP_E2E_JUNIT")
	if path == "" {
		path = "e2e-report.xml"
	}
	if err := report.Default.WriteJUnit(path); err != nil {
		fmt.Fprintf(os.Stderr, "write JUnit report: %v\n", err)
	} else {
		cases := report.Default.Cases()
		failed, skipped := 0, 0
		for _, c := range cases {
			switch {
			case c.Failed:
				failed++
			case c.Skipped:
				skipped++
			}
		}
		fmt.Printf("\nE2E summary: %d tests, %d passed, %d failed, %d skipped -> %s\n",
			len(cases), len(cases)-failed-skipped, failed, skipped, path)
	}

	driftPath := os.Getenv("UCRP_E2E_DRIFT_REPORT")
	if driftPath == "" {
		driftPath = "e2e-contract-drift.md"
	}
	if err := report.Drifts.WriteMarkdown(driftPath); err != nil {
		fmt.Fprintf(os.Stderr, "write contract drift report: %v\n", err)
	} else {
		fmt.Printf("Contract drift entries: %d -> %s\n", len(report.Drifts.All()), driftPath)
	}
	os.Exit(code)
}
