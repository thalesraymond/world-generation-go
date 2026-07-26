package exporter

import (
	"fmt"
	"strings"
)

type field struct {
	Key   string
	Value string
}

func frontmatter(fields []field) string {
	if len(fields) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("---\n")
	for _, f := range fields {
		fmt.Fprintf(&b, "%s: %s\n", f.Key, quoteIfNeeded(f.Value))
	}
	b.WriteString("---\n")
	return b.String()
}

func quoteIfNeeded(v string) string {
	if strings.ContainsAny(v, ":#{}[]&*!|>'\"%@`") || v == "" || v == "true" || v == "false" || v == "null" {
		return fmt.Sprintf("%q", v)
	}
	return v
}
