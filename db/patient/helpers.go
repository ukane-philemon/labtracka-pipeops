package patient

import "strings"

// mapKey chain MongoDB field keys.
func mapKey(fields ...string) string {
	var field string
	for _, f := range fields {
		field += "." + f
	}
	return strings.Trim(field, ".")
}
