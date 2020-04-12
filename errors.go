package firephone

type ErrorCode string

const (
	ErrorCodeInvalidAPIKey           = "INVALID_KEY"
	ErrorCodeInvalidCaptchaKey       = "INVALID_CAPTCHA_KEY"
	ErrorCodeInvalidConfirmationCode = "INVALID_CONFIRMATION_CODE"
)

type Err struct {
	Message   string
	ErrorCode ErrorCode
}

func (m *Err) Error() string {
	return m.Message
}

func IsErrorType(err error, errCode ErrorCode) bool {
	serr, ok := err.(*Err)
	if ok && serr.ErrorCode == errCode {
		return true
	}
	return false
}
