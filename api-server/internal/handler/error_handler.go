package handler

import (
	"errors"
	"log"

	apperrors "wuxie-api/pkg/errors"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// respondWithError 根据错误类型返回对应的 HTTP 响应
func respondWithError(c *gin.Context, err error) {
	// 认证错误 → 401
	if apperrors.IsAuthError(err) {
		response.Unauthorized(c, err.Error())
		return
	}

	// 权限错误 → 403
	if apperrors.IsAccessDenied(err) {
		response.Forbidden(c, err.Error())
		return
	}

	// 账号封禁 → 403
	if errors.Is(err, apperrors.ErrAccountSuspended) {
		response.Forbidden(c, err.Error())
		return
	}

	// 参数错误 → 400
	if errors.Is(err, apperrors.ErrInvalidParams) {
		response.BadRequest(c, err.Error())
		return
	}

	// 未找到 → 404
	if apperrors.IsNotFound(err) {
		response.NotFound(c, err.Error())
		return
	}

	// 其他错误 → 500（不暴露内部信息）
	log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	response.InternalError(c, "internal server error")
}
