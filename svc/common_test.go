package svc

import (
	"testing"
)

func TestHashPass(t *testing.T) {
	cases := []struct {
		in, name   string
		shouldFail bool
	}{
		{
			in:         "foo",
			name:       "non-empty password",
			shouldFail: false,
		},
		{
			in:         "",
			name:       "empty password",
			shouldFail: true,
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			out, err := HashPass(v.in)
			if err != nil && !v.shouldFail {
				t.Errorf("Shouldn't have failed: Case %s, Error: %s", v.name, err)
			}
			if out == "" && v.in != "" {
				t.Errorf("Got empty out for case %s input %s", v.name, v.in)
			}
		})
	}
}
