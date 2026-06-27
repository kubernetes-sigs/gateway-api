/*
Copyright 2025 The Kubernetes Authors.

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

package logging

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

type mockArrayEncoder struct {
	zapcore.PrimitiveArrayEncoder
	strings []string
}

func (m *mockArrayEncoder) AppendString(s string) {
	m.strings = append(m.strings, s)
}

func TestCustomLevelEncoder(t *testing.T) {
	tests := []struct {
		name     string
		level    zapcore.Level
		expected string
	}{
		{
			name:     "Standard Info (0)",
			level:    zapcore.InfoLevel, // 0
			expected: "info",
		},
		{
			name:     "Standard Warn (1)",
			level:    zapcore.WarnLevel, // 1
			expected: "warn",
		},
		{
			name:     "Standard Error (2)",
			level:    zapcore.ErrorLevel, // 2
			expected: "error",
		},
		{
			name:     "V(1) (-1)",
			level:    zapcore.Level(-1),
			expected: "info",
		},
		{
			name:     "V(2) Default (-2)",
			level:    zapcore.Level(-2),
			expected: "info",
		},
		{
			name:     "Verbose (-3)",
			level:    zapcore.Level(-3),
			expected: "info",
		},
		{
			name:     "Debug (-4)",
			level:    zapcore.Level(-4),
			expected: "debug",
		},
		{
			name:     "Trace (-5)",
			level:    zapcore.Level(-5),
			expected: "trace",
		},
		{
			name:     "Extremely Verbose (-6)",
			level:    zapcore.Level(-6),
			expected: "trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := &mockArrayEncoder{}
			customLevelEncoder(tt.level, enc)
			if len(enc.strings) != 1 {
				t.Fatalf("Expected 1 string appended, got %d", len(enc.strings))
			}
			if enc.strings[0] != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, enc.strings[0])
			}
		})
	}
}
