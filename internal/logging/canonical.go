package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// CanonicalLogger collects key=value pairs and emits them as a single log line.
type CanonicalLogger struct {
	fields []field
	w      io.Writer
}

type field struct {
	key   string
	value interface{}
}

// NewCanonicalLogger creates a new logger that writes to stderr.
func NewCanonicalLogger(w io.Writer) *CanonicalLogger {
	if w == nil {
		w = os.Stderr
	}
	return &CanonicalLogger{w: w}
}

// Add adds a single key-value pair.
func (l *CanonicalLogger) Add(key string, value interface{}) {
	l.fields = append(l.fields, field{key: key, value: value})
}

// AddMany adds multiple key-value pairs preserving insertion order.
func (l *CanonicalLogger) AddMany(pairs map[string]interface{}) {
	for k, v := range pairs {
		l.fields = append(l.fields, field{key: k, value: v})
	}
}

// Emit writes all collected fields as a single canonical log line and resets.
func (l *CanonicalLogger) Emit() {
	parts := make([]string, 0, len(l.fields))
	for _, f := range l.fields {
		parts = append(parts, f.key+"="+formatValue(f.value))
	}
	fmt.Fprintln(l.w, strings.Join(parts, " "))
	l.fields = l.fields[:0]
}

func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	s := fmt.Sprintf("%v", v)
	if strings.Contains(s, " ") || s == "" {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		return `"` + s + `"`
	}
	return s
}
