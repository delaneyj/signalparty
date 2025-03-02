package templates

import (
	"strconv"
	"strings"
)

func prefixedStrings(prefix string, count int) string {
	var sb strings.Builder
	for i := 0; i < count; i++ {
		sb.WriteString(prefix)
		sb.WriteString(strconv.Itoa(i))
		if i < count-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}
