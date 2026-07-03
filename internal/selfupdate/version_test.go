package selfupdate

import "testing"

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in   string
		want *Version
	}{
		{"1.2.3", &Version{1, 2, 3}},
		{"v1.2.3", &Version{1, 2, 3}},
		{"V2.0.0", &Version{2, 0, 0}},
		{"1.2.3-rc1", &Version{1, 2, 3}},
		{"1.2.3+build.5", &Version{1, 2, 3}},
		{"1.2", &Version{1, 2, 0}},
		{"1", &Version{1, 0, 0}},
		{" 0.1.0 ", &Version{0, 1, 0}},
		{"dev", nil},
		{"", nil},
		{"1.x.3", nil},
	}
	for _, c := range cases {
		got := ParseVersion(c.in)
		if c.want == nil {
			if got != nil {
				t.Errorf("ParseVersion(%q) = %+v, want nil", c.in, got)
			}
			continue
		}
		if got == nil || *got != *c.want {
			t.Errorf("ParseVersion(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"1.2.4", "1.2.3", true},
		{"1.3.0", "1.2.9", true},
		{"2.0.0", "1.9.9", true},
		{"1.2.3", "1.2.3", false},
		{"1.2.2", "1.2.3", false},
		{"v0.2.0", "0.1.0", true},
		// unparseable current ("dev") → any real version is newer
		{"0.1.0", "dev", true},
		// unparseable latest → never newer
		{"garbage", "1.0.0", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.latest, c.current); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"v1.2.3": "1.2.3",
		"1.2.3":  "1.2.3",
		" v0.1 ": "0.1",
	} {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
