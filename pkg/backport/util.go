package backport

import "strings"

func Cut(s, sep string) (before, after string, found bool) {
	sp := strings.SplitN(s, sep, 2)
	if len(sp) == 2 {
		return sp[0], sp[1], true
	}
	return s, "", false
}
