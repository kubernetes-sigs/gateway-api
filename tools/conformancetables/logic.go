/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import "strings"

const httpRoute = "HTTPRoute"

func isUpper(b byte) bool { return b >= 'A' && b <= 'Z' }
func isLower(b byte) bool { return b >= 'a' && b <= 'z' }
func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// ProcessFeatureName splits a CamelCase feature name into space-separated words.
//
// It is based on the regex used by the original Python script:
//
//	HTTPRoute|[A-Z]+(?=[A-Z][a-z])|[A-Z][a-z]+|[A-Z\d]+
//
// with one fix: the last alternative no longer swallows an uppercase letter
// that starts a capitalized word, so "303Redirect" becomes "303 Redirect"
// rather than "303R". The alternatives are tried in order at each position,
// and characters that match none of them are dropped. Go's regexp has no
// lookahead, so the matching is done by hand.
func ProcessFeatureName(s string) string {
	var tokens []string
	i := 0
	for i < len(s) {
		switch {
		case strings.HasPrefix(s[i:], httpRoute):
			tokens = append(tokens, httpRoute)
			i += len(httpRoute)
		default:
			if end, ok := upperRunBeforePair(s, i); ok {
				tokens = append(tokens, s[i:end])
				i = end
			} else if end, ok := capitalizedWord(s, i); ok {
				tokens = append(tokens, s[i:end])
				i = end
			} else if end, ok := upperOrDigitRun(s, i); ok {
				tokens = append(tokens, s[i:end])
				i = end
			} else {
				i++ // no rule matched; the Python version drops this character
			}
		}
	}
	return strings.Join(tokens, " ")
}

// upperRunBeforePair implements [A-Z]+(?=[A-Z][a-z]): a greedy run of
// uppercase letters that backs off until the next two characters are an
// uppercase letter followed by a lowercase one.
func upperRunBeforePair(s string, i int) (int, bool) {
	j := i
	for j < len(s) && isUpper(s[j]) {
		j++
	}
	for e := j - 1; e > i; e-- {
		if e+1 < len(s) && isUpper(s[e]) && isLower(s[e+1]) {
			return e, true
		}
	}
	return 0, false
}

// capitalizedWord implements [A-Z][a-z]+.
func capitalizedWord(s string, i int) (int, bool) {
	if !isUpper(s[i]) {
		return 0, false
	}
	j := i + 1
	for j < len(s) && isLower(s[j]) {
		j++
	}
	return j, j > i+1
}

// upperOrDigitRun implements [A-Z\d]+, except that it stops before an
// uppercase letter that starts a capitalized word (an uppercase letter
// followed by a lowercase one). "303Redirect" therefore yields "303" and
// leaves "Redirect" for the next token.
func upperOrDigitRun(s string, i int) (int, bool) {
	j := i
	for j < len(s) && (isUpper(s[j]) || isDigit(s[j])) {
		if j > i && isUpper(s[j]) && j+1 < len(s) && isLower(s[j+1]) {
			break
		}
		j++
	}
	return j, j > i
}
