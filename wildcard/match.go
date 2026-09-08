// Copyright (c) 2015-2023 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package wildcard

import "strings"

// MatchSimple - finds whether the text matches/satisfies the pattern string.
// supports '*' wildcard in the pattern and ? for single characters.
// Unlike Match, reaching '?' after name is exhausted accepts the rest of the
// pattern. For example, both "a?" and "a?b" match "a".
func MatchSimple(pattern, name string) bool {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	return deepMatchRune(name, pattern, true)
}

// Match -  finds whether the text matches/satisfies the pattern string.
// supports  '*' and '?' wildcards in the pattern string.
// unlike path.Match(), considers a path as a flat name space while matching the pattern.
// The difference is illustrated in the example here https://play.golang.org/p/Ega9qgD4Qz .
func Match(pattern, name string) (matched bool) {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	// Do an extended wildcard '*' and '?' match.
	return deepMatchRune(name, pattern, false)
}

// Has returns true if the input pattern has a wildcard (pattern).
func Has(pattern string) bool {
	return strings.ContainsAny(pattern, "*?")
}

// Keep one backtracking point instead of recursively exploring both
// alternatives for every '*'.
func deepMatchRune(str, pattern string, simple bool) bool {
	var s, p int
	// Position of the '*' to resume from, and how much of str it has consumed.
	star, mark := -1, 0
	for s < len(str) || p < len(pattern) {
		if p < len(pattern) {
			switch pattern[p] {
			case '*':
				star, mark = p, s
				p++
				if p == len(pattern) {
					return true
				}
				if simple {
					// Before a later '*' replaces this one, check whether its
					// star-free segment can reach '?' with the name exhausted.
					for end := p; end < len(pattern) && pattern[end] != '*' && end-p <= len(str)-s; end++ {
						if pattern[end] == '?' && matchFixedSuffix(str[s:], pattern[p:end]) {
							return true
						}
					}
				}
				continue
			case '?':
				if simple && s == len(str) {
					return true
				}
				if s < len(str) {
					s++
					p++
					continue
				}
			default:
				if s < len(str) && pattern[p] == str[s] {
					s++
					p++
					continue
				}
			}
		}
		if star < 0 {
			return false
		}
		// Let the last '*' swallow one more byte and retry from there.
		mark++
		if mark > len(str) {
			return false
		}
		s, p = mark, star+1
	}
	return true
}

// matchFixedSuffix checks a star-free pattern already known to fit in str.
func matchFixedSuffix(str, pattern string) bool {
	str = str[len(str)-len(pattern):]
	for i := len(pattern) - 1; i >= 0; i-- {
		if pattern[i] != '?' && pattern[i] != str[i] {
			return false
		}
	}
	return true
}

// MatchAsPatternPrefix matches text as a prefix of the given pattern. Examples:
//
//	| Pattern | Text    | Match Result |
//	====================================
//	| abc*    | ab      | True         |
//	| abc*    | abd     | False        |
//	| abc*c   | abcd    | True         |
//	| ab*??d  | abxxc   | True         |
//	| ab*??d  | abxc    | True         |
//	| ab??d   | abxc    | True         |
//	| ab??d   | abc     | True         |
//	| ab??d   | abcxdd  | False        |
//
// This function is only useful in some special situations.
func MatchAsPatternPrefix(pattern, text string) bool {
	for i := 0; i < len(text) && i < len(pattern); i++ {
		if pattern[i] == '*' {
			return true
		}
		if pattern[i] == '?' {
			continue
		}
		if pattern[i] != text[i] {
			return false
		}
	}
	return len(text) <= len(pattern)
}
