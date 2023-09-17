package main

import (
	"strings"
)

func toSnakeCase(s string) string {
	lowercased := strings.ToLower(s)
	return strings.Replace(lowercased, " ", "_", -1)
}