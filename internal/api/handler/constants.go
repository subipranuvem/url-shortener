package handler

const (
	PathParamCode = "code"

	ErrCodeNotFound      = "not_found"
	ErrCodeInternalError = "internal_error"
	ErrCodeAliasTaken    = "alias_already_taken"
	ErrCodeURLInactive   = "url_inactive"

	RespKeyCode       = "code"
	RespKeyShortURL   = "short_url"
	RespKeyClickCount = "click_count"
	RespKeyError      = "error"
)
