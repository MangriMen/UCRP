package seed

import "fmt"

type rng struct {
	state uint64
}

func newRNG(seed uint64) *rng {
	return &rng{state: seed + 0x9e3779b97f4a7c15}
}

func (r *rng) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (r *rng) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() % uint64(n))
}

func (r *rng) chance(percent int) bool {
	return r.intn(100) < percent
}

const demoMailDomain = "demo" + "." + "ucrp" + "." + "test"

func demoPhone(index int) string {
	return fmt.Sprintf("79%09d", 100000000+index)
}

func demoEmail(index int) string {
	return fmt.Sprintf("guest%03d@%s", index, demoMailDomain)
}

func demoCard(index int) string {
	return fmt.Sprintf("4000%012d", index)
}
