package middleware

import "context"

func GetUserIDFromCtx(ctx context.Context) uint {
	v, _ := ctx.Value(ContextUserID).(uint)
	return v
}

func GetRoleFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ContextRole).(string)
	return v
}
