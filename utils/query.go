package utils

import (
	"strconv"
	"strings"
)

func ParseUintSlice(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []uint
	for _, p := range parts {
		if val, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64); err == nil {
			result = append(result, uint(val))
		}
	}
	return result
}
