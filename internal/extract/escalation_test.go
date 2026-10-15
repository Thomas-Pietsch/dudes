package extract

import "testing"

func TestFindEscalation(t *testing.T) {
	cases := []struct {
		in   string
		kind string
		ok   bool
	}{
		{"working…\nBLOCKED: no route for edit — TA would own it\n", "BLOCKED", true},
		{"  CONTRACT: [[TA:API-13]] — 409 vs 400 contradicts AC-3", "CONTRACT", true},
		{"report `BLOCKED: <what is missing>` instead of inventing", "", false},
		{"===BEGIN ARTIFACT===\nBLOCKED: T2 field hypothesis\n===END ARTIFACT===", "", false},
		{"done, all ACs covered", "", false},
	}
	for _, c := range cases {
		e, ok := FindEscalation(c.in)
		if ok != c.ok || e.Kind != c.kind {
			t.Fatalf("%q → %+v %v", c.in, e, ok)
		}
	}
}
