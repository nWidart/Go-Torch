package app

import "testing"

func TestStrconvItoa(t *testing.T) {
	cases := map[int]string{
		0:    "0",
		1:    "1",
		-1:   "-1",
		42:   "42",
		1234: "1234",
	}
	for n, want := range cases {
		if got := strconvItoa(n); got != want {
			t.Fatalf("itoa(%d) = %q; want %q", n, got, want)
		}
	}
}
