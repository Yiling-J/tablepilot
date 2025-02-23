package utils

import "github.com/gin-gonic/gin"

const ctxUserKey = "ctxUserKey"

func WithUserId(ctx *gin.Context, userId int) {
	ctx.Set(ctxUserKey, userId)
}

func UserId(ctx *gin.Context) int {
	raw, exists := ctx.Get(ctxUserKey)
	if !exists {
		return 0
	}
	return raw.(int)
}
