package geosite

import (
	"strings"
)

func ParseExpr(expr string) (string, []string) {
	parts := strings.Split(expr, "@")
	if len(parts) == 0 {
		return "", nil
	}
	base := strings.TrimSpace(parts[0])
	extra := parts[1:]
	if len(base) == 0 {
		return "", nil
	}
	attrs := make([]string, 0)
	for _, attr := range extra {
		attr = strings.ToLower(strings.TrimSpace(attr))
		if len(attr) == 0 {
			continue
		}
		attrs = append(attrs, attr)
	}
	if len(attrs) == 0 {
		return base, nil
	}
	return base, attrs
}

func (attrs RecordAttr) ContainAll(targets []string) bool {
	miss := false
	for _, target := range targets {
		match := false
		for attr := range attrs {
			if strings.EqualFold(attr, target) {
				match = true
				break
			}
		}
		if !match {
			miss = true
			break
		}
	}
	return !miss
}

// Filters geosite by given expression,
// @attr expression supported.
// Returns nil on no match.
func (col Collection) Select(expr string) RecordList {
	base, attrs := ParseExpr(expr)
	if len(base) == 0 {
		return nil
	}
	baseMatched := col[strings.ToLower(base)]
	if len(baseMatched) == 0 {
		return nil
	}
	if len(attrs) == 0 {
		return baseMatched
	}
	filtered := make(RecordList, 0)
	for _, record := range baseMatched {
		if record.Attr.ContainAll(attrs) {
			filtered = append(filtered, record)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}
