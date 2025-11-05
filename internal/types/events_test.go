package types

import "testing"

func TestEventKindString(t *testing.T) {
	cases := []struct {
		in   EventKind
		want string
	}{
		{EventUnknown, "Unknown"},
		{EventMapStart, "MapStart"},
		{EventMapEnd, "MapEnd"},
		{EventBagInit, "BagInit"},
		{EventBagMod, "BagMod"},
		{EventKind(999), "Unknown"},
	}
	for _, c := range cases {
		if got := c.in.String(); got != c.want {
			t.Fatalf("%v.String() = %q; want %q", c.in, got, c.want)
		}
	}
}
