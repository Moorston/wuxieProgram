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
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenRevoked   = errors.New("token has been revoked")
	ErrInvalidRefresh = errors.New("invalid refresh token")
	ErrUserNotFound   = errors.New("user not found")
	ErrMissingAuth    = errors.New("missing authorization header")
	ErrInvalidFormat  = errors.New("invalid authorization format")
	ErrInvalidUserID  = errors.New("invalid user identity")
)

// 打卡相关错误
var (
	ErrCheckinNotFound = errors.New("checkin not found")
	ErrNotCheckinOwner = errors.New("access denied: not checkin owner")
)

// 资源相关错误
var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrNotResourceOwner = errors.New("access denied: not resource owner")
)

// 感悟相关错误
var (
	ErrInsightNotFound = errors.New("insight not found")
	ErrNotInsightOwner = errors.New("access denied: not insight owner")
)

// 训练相关错误
var (
	ErrPlanNotFound   = errors.New("training plan not found")
	ErrTaskOutOfRange = errors.New("task index out of range")
)

// 通知相关错误
var (
	ErrNotificationNotFound = errors.New("notification not found")
)

// 管理后台错误
var (
	ErrAdminNotConfigured = errors.New("admin not configured")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTooManyIDs         = errors.New("too many ids")
)

// IsNotFound 判断是否为"未找到"类错误
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrCheckinNotFound) ||
		errors.Is(err, ErrResourceNotFound) ||
		errors.Is(err, ErrInsightNotFound) ||
		errors.Is(err, ErrPlanNotFound) ||
		errors.Is(err, ErrNotificationNotFound)
}

// IsAccessDenied 判断是否为"权限不足"类错误
func IsAccessDenied(err error) bool {
	return errors.Is(err, ErrAccessDenied) ||
		errors.Is(err, ErrNotCheckinOwner) ||
		errors.Is(err, ErrNotResourceOwner) ||
		errors.Is(err, ErrNotInsightOwner)
}

// IsAuthError 判断是否为认证类错误
func IsAuthError(err error) bool {
	return errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrTokenRevoked) ||
		errors.Is(err, ErrInvalidRefresh) ||
		errors.Is(err, ErrMissingAuth) ||
		errors.Is(err, ErrInvalidFormat) ||
		errors.Is(err, ErrInvalidUserID) ||
		errors.Is(err, ErrInvalidCredentials)
}
