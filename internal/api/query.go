package api

import (
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"strconv"
	"strings"
	"time"
)

// resolveIntQuery retrieves and parses an integer query parameter from the context.
// If the parameter is present and valid, it returns a pointer to the parsed integer.
// Otherwise, it returns nil.
func resolveIntQuery(ctx huma.Context, paramName string) (*int, error) {
	if v := ctx.Query(paramName); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %v", paramName, err)
		}
		tmp := int(parsed)
		return &tmp, nil
	}
	return nil, nil
}

// resolveStringQuery retrieves a string query parameter from the context.
// If the parameter is present, it returns a pointer to the string value.
// Otherwise, it returns nil.
func resolveStringQuery(ctx huma.Context, paramName string) (*string, error) {
	if v := ctx.Query(paramName); v != "" {
		return &v, nil
	}
	return nil, nil
}

// resolveStringSliceQuery retrieves a comma-separated string slice query parameter from the context.
// If the parameter is present, it splits the string into a slice of strings and returns it.
// Otherwise, it returns nil.
func resolveStringSliceQuery(ctx huma.Context, paramName string) ([]string, error) {
	if v := ctx.Query(paramName); v != "" {
		items := strings.Split(v, ",")
		return items, nil
	}
	return nil, nil
}

// resolveTimeQuery retrieves and parses a time query parameter from the context.
// If the parameter is present and valid (in RFC3339 format), it returns a pointer to the parsed time.
// Otherwise, it returns nil.
func resolveTimeQuery(ctx huma.Context, paramName string) (*time.Time, error) {
	if v := ctx.Query(paramName); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %v", paramName, err)
		}
		return &parsed, nil
	}
	return nil, nil
}
