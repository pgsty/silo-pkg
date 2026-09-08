// Copyright (c) 2015-2026 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package wildcard

import (
	"strings"
	"testing"
)

// oldDeepMatchRune is the recursive matcher deepMatchRune replaced, kept here so
// the rewrite can be proven equivalent rather than asserted to be.
func oldDeepMatchRune(str, pattern string, simple bool) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		default:
			if len(str) == 0 || str[0] != pattern[0] {
				return false
			}
		case '?':
			if len(str) == 0 {
				return simple
			}
		case '*':
			return len(pattern) == 1 ||
				oldDeepMatchRune(str, pattern[1:], simple) ||
				(len(str) > 0 && oldDeepMatchRune(str[1:], pattern, simple))
		}
		str = str[1:]
		pattern = pattern[1:]
	}
	return len(str) == 0 && len(pattern) == 0
}

func gen(alphabet string, maxLen int) []string {
	out := []string{""}
	cur := []string{""}
	for range maxLen {
		var next []string
		for _, s := range cur {
			for _, c := range alphabet {
				next = append(next, s+string(c))
			}
		}
		out = append(out, next...)
		cur = next
	}
	return out
}

// Keep the previous matcher as the compatibility oracle for both modes.
func TestDeepMatchEquivalenceExhaustive(t *testing.T) {
	pats := gen("ab*?", 5)
	names := gen("ab", 5)
	var n int
	for _, p := range pats {
		for _, name := range names {
			if got, want := Match(p, name), oldMatch(p, name); got != want {
				t.Fatalf("Match(%q, %q) = %v, old = %v", p, name, got, want)
			}
			if got, want := MatchSimple(p, name), oldMatchSimple(p, name); got != want {
				t.Fatalf("MatchSimple(%q, %q) = %v, old = %v", p, name, got, want)
			}
			n += 2
		}
	}
	t.Logf("%d pattern/name/mode combinations agree", n)
}

// Same for the exported entry points, over a colon-bearing alphabet closer to
// policy actions and resource ARNs.
func TestMatchEquivalenceExhaustive(t *testing.T) {
	pats := gen("a:*?", 4)
	names := gen("a:/", 4)
	var n int
	for _, p := range pats {
		for _, name := range names {
			if got, want := Match(p, name), oldMatch(p, name); got != want {
				t.Fatalf("Match(%q, %q) = %v, old = %v", p, name, got, want)
			}
			if got, want := MatchSimple(p, name), oldMatchSimple(p, name); got != want {
				t.Fatalf("MatchSimple(%q, %q) = %v, old = %v", p, name, got, want)
			}
			n += 2
		}
	}
	t.Logf("%d exported-entry-point combinations agree", n)
}

func oldMatch(pattern, name string) bool {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	return oldDeepMatchRune(name, pattern, false)
}

func oldMatchSimple(pattern, name string) bool {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	return oldDeepMatchRune(name, pattern, true)
}

func TestMatchManyStars(t *testing.T) {
	name := "admin:ServerInfo"
	for _, stars := range []int{8, 16, 64, 256} {
		pattern := strings.Repeat("*", stars) + "X"
		if Match(pattern, name) {
			t.Fatalf("pattern %d stars + X should not match %q", stars, name)
		}
		if MatchSimple(pattern, name) {
			t.Fatalf("simple pattern %d stars + X should not match %q", stars, name)
		}
	}
	// Interleaved stars are the harder shape.
	pattern := strings.Repeat("*a", 32) + "X"
	name = strings.Repeat("a", 128)
	if Match(pattern, name) {
		t.Errorf("pattern %q should not match %q", pattern, name)
	}
	if MatchSimple(pattern, name) {
		t.Errorf("simple pattern %q should not match %q", pattern, name)
	}
}

func BenchmarkMatchStarBacktracking(b *testing.B) {
	pattern, name := "********X", "admin:ServerInfo"
	for _, matcher := range []struct {
		name  string
		match func(string, string) bool
	}{
		{"current", Match},
		{"previous", oldMatch},
	} {
		b.Run(matcher.name, func(b *testing.B) {
			for b.Loop() {
				matcher.match(pattern, name)
			}
		})
	}
}

// Cover optional question marks and long near misses. Star counts stay low:
// the previous matcher is exponential in them.
func BenchmarkMatchSimpleQuestionMarks(b *testing.B) {
	cases := []struct {
		name    string
		pattern string
		text    string
	}{
		{"star-free", strings.Repeat("a?", 16), strings.Repeat("aa", 16)},
		{"all-marks", strings.Repeat("?", 32), strings.Repeat("a", 16)},
		{"leading-star", "*" + strings.Repeat("a?", 8), strings.Repeat("a", 24)},
		{"no-mark", "arn:aws:s3:::*", "arn:aws:s3:::bucket/object"},
		{"trailing-star-1024", "bucket/*", "bucket/" + strings.Repeat("a", 1024)},
		{"near-miss-64", "*" + strings.Repeat("a?", 32) + "X", strings.Repeat("a", 64) + "b"},
		{"near-miss-256", "*" + strings.Repeat("a?", 128) + "X", strings.Repeat("a", 256) + "b"},
		{"near-miss-1024", "*" + strings.Repeat("a?", 512) + "X", strings.Repeat("a", 1024) + "b"},
	}
	for _, c := range cases {
		b.Run("new/"+c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = MatchSimple(c.pattern, c.text)
			}
		})
		b.Run("old/"+c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = oldMatchSimple(c.pattern, c.text)
			}
		})
	}
}

func TestMatchSimpleExhaustedName(t *testing.T) {
	for _, tc := range []struct {
		pattern, name string
		want          bool
	}{
		{"a?b", "a", true},
		{"a?b", "ac", false},
		{"?suffix", "", true},
		{"*?*a", "a", true},
		{"*a?b", "ca", true},
		{"*a?b", "cb", false},
		{"?", "é", false},
		{"??", "é", true},
	} {
		if got := MatchSimple(tc.pattern, tc.name); got != tc.want {
			t.Errorf("MatchSimple(%q, %q) = %v, want %v", tc.pattern, tc.name, got, tc.want)
		}
	}
}

func FuzzDeepMatchEquivalence(f *testing.F) {
	for _, s := range []string{"", "*", "?", "a*b", "admin:*", "**?", "a", "*a*a*b"} {
		f.Add(s, "admin:Heal")
	}
	f.Add("*?*a", "a")
	f.Add("?", "\xff")
	f.Add("??", "é")
	f.Add("对象/*", "对象/路径")
	f.Fuzz(func(t *testing.T, pattern, name string) {
		// Bound the old implementation's exponential blowup so the fuzzer
		// compares results instead of timing out on the bug being fixed.
		if strings.Count(pattern, "*") > 4 || len(pattern) > 24 || len(name) > 24 {
			t.Skip()
		}
		if got, want := Match(pattern, name), oldMatch(pattern, name); got != want {
			t.Fatalf("Match(%q, %q) = %v, old = %v", pattern, name, got, want)
		}
		if got, want := MatchSimple(pattern, name), oldMatchSimple(pattern, name); got != want {
			t.Fatalf("MatchSimple(%q, %q) = %v, old = %v", pattern, name, got, want)
		}
	})
}
