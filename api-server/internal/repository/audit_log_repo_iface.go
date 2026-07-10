package repository

import (
	"context"

	"wuxie-api/internal/model"
)

// AuditLogInterface 审计日志仓库接口
type AuditLogInterface interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, page, pageSize int) ([]*model.AuditLog, int64, error)
}

var _ AuditLogInterface = (*AuditLogRepo)(nil)
