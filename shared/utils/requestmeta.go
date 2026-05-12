package utils

import "context"

func StringFromContext(ctx context.Context, key any) string {
	v := ctx.Value(key)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
