/*
Copyright 2026 The Kubernetes Authors.

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

package http //nolint:revive

import (
	"fmt"
	"net/http"
	"strings"
)

// CookieInfo holds the parsed name, value, and attributes of a Set-Cookie header.
type CookieInfo struct {
	Name       string
	Value      string
	Attributes map[string]string
}

// ExtractResponseCookie parses the first Set-Cookie header from the response and
// returns the cookie name, value, and a map of its attributes (e.g. "Max-Age", "Expires").
func ExtractResponseCookie(h http.Header) (*CookieInfo, error) {
	values := h.Values("Set-Cookie")
	if len(values) == 0 {
		return nil, fmt.Errorf("no Set-Cookie header found in response")
	}

	raw := strings.TrimSpace(values[0])
	if raw == "" {
		return nil, fmt.Errorf("empty Set-Cookie header")
	}

	parts := strings.Split(raw, ";")
	if len(parts) == 0 {
		return nil, fmt.Errorf("malformed Set-Cookie header: %q", raw)
	}

	// First part is cookie-name=value.
	pair := strings.TrimSpace(parts[0])
	nv := strings.SplitN(pair, "=", 2)
	if len(nv) != 2 {
		return nil, fmt.Errorf("malformed Set-Cookie header (no name=value): %q", raw)
	}

	name := strings.TrimSpace(nv[0])
	value := strings.TrimSpace(nv[1])
	if name == "" || value == "" {
		return nil, fmt.Errorf("malformed Set-Cookie header (empty name or value): %q", raw)
	}

	attrs := make(map[string]string)
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := ""
		if len(kv) == 2 {
			val = strings.TrimSpace(kv[1])
		}
		attrs[key] = val
	}

	return &CookieInfo{Name: name, Value: value, Attributes: attrs}, nil
}

// HasAttribute reports whether the cookie has the given attribute (case-insensitive).
func (c *CookieInfo) HasAttribute(attr string) bool {
	_, ok := c.Attributes[strings.ToLower(attr)]
	return ok
}
