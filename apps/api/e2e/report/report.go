package report

import (
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type Case struct {
	Name     string
	Class    string
	Duration time.Duration
	Failed   bool
	Skipped  bool
	Message  string
}

type Recorder struct {
	mu    sync.Mutex
	cases []Case
}

var Default = &Recorder{}

func (r *Recorder) Add(c Case) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cases = append(r.cases, c)
}

func (r *Recorder) Cases() []Case {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Case, len(r.cases))
	copy(out, r.cases)
	return out
}

type junitTestsuite struct {
	XMLName  xml.Name        `xml:"testsuite"`
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Skipped  int             `xml:"skipped,attr"`
	Time     string          `xml:"time,attr"`
	Cases    []junitTestcase `xml:"testcase"`
}

type junitTestcase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	Skipped   *junitSkipped `xml:"skipped,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr"`
}

func (r *Recorder) WriteJUnit(path string) error {
	cases := r.Cases()
	sort.SliceStable(cases, func(i, j int) bool {
		if cases[i].Class != cases[j].Class {
			return cases[i].Class < cases[j].Class
		}
		return cases[i].Name < cases[j].Name
	})
	suite := junitTestsuite{Name: "ucrp-e2e", Tests: len(cases)}
	var total time.Duration
	for _, c := range cases {
		total += c.Duration
		tc := junitTestcase{
			Name:      c.Name,
			Classname: c.Class,
			Time:      fmt.Sprintf("%.3f", c.Duration.Seconds()),
		}
		switch {
		case c.Failed:
			suite.Failures++
			tc.Failure = &junitFailure{Message: c.Message, Body: c.Message}
		case c.Skipped:
			suite.Skipped++
			tc.Skipped = &junitSkipped{Message: c.Message}
		}
		suite.Cases = append(suite.Cases, tc)
	}
	suite.Time = fmt.Sprintf("%.3f", total.Seconds())
	data, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal junit: %w", err)
	}
	payload := append([]byte(xml.Header), data...)
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write junit %s: %w", path, err)
	}
	return nil
}
