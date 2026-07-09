package handler

import (
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// getUserID 从上下文中提取并验证用户ID，失败时自动返回错误并中止请求
func getUserID(c *gin.Context) (primitive.ObjectID, bool) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user identity")
		return primitive.NilObjectID, false
	}

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.Unauthorized(c, "invalid user identity")
		return primitive.NilObjectID, false
	}

	return oid, true
}

// getObjectID 从路径参数中提取并验证ObjectID
func getObjectID(c *gin.Context, param string) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param(param))
	if err != nil {
		response.BadRequest(c, "invalid "+param)
		return primitive.NilObjectID, false
	}
	return id, true
}