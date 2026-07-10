package errors

import "errors"

// 通用业务错误
var (
	ErrNotFound         = errors.New("not found")
	ErrAccessDenied     = errors.New("access denied")
	ErrAccountSuspended = errors.New("account has been suspended")
	ErrInvalidParams    = errors.New("invalid parameters")
	ErrInternal         = errors.New("internal server error")
)

// 认证相关错误
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenRevoked     = errors.New("token has been revoked")
	ErrInvalidRefresh   = errors.New("invalid refresh token")
	ErrUserNotFound     = errors.New("user not found")
)

// 打卡相关错误
var (
	ErrCheckinNotFound  = errors.New("checkin not found")
	ErrNotCheckinOwner  = errors.New("access denied: not checkin owner")
)

// 资源相关错误
var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrNotResourceOwner = errors.New("access denied: not resource owner")
)

// 感悟相关错误
var (
	ErrInsightNotFound  = errors.New("insight not found")
	ErrNotInsightOwner  = errors.New("access denied: not insight owner")
)