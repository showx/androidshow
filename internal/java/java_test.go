package java

import "testing"

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in     string
		want   int
		wantOK bool
	}{
		{`java version "1.8.0_392"`, 8, true},
		{`openjdk version "11.0.22" 2024-01-16`, 11, true},
		{`openjdk version "17.0.9" 2023-10-17`, 17, true},
		{`openjdk version "21.0.2" 2024-01-16 LTS`, 21, true},
		{`not a java version`, 0, false},
	}
	for _, c := range cases {
		got, ok := ParseVersion(c.in)
		if ok != c.wantOK || got != c.want {
			t.Fatalf("ParseVersion(%q) = (%d, %v), want (%d, %v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}
