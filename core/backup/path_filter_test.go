package backup

import "testing"

func TestShouldSkipBackupPath(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{rel: ".", want: false},
		{rel: "skills/a/skill.md", want: false},
		{rel: "meta/123.json", want: false},
		{rel: "config.json", want: false},
		{rel: "cache", want: true},
		{rel: "cache/tmp.bin", want: true},
		{rel: "./cache/tmp.bin", want: true},
		{rel: ".git", want: true},
		{rel: ".git/index", want: true},
		{rel: "./.git/objects/xx", want: true},
	}
	for _, tc := range cases {
		got := ShouldSkipBackupPath(tc.rel)
		if got != tc.want {
			t.Fatalf("ShouldSkipBackupPath(%q)=%v, want %v", tc.rel, got, tc.want)
		}
	}
}
