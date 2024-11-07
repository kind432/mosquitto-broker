package consts

// http code 400
const (
	ErrEmailAlreadyInUse        = "email already in use"
	ErrAtoi                     = "string to int error"
	ErrTopicIsExist             = "topic is already exist"
	ErrTimeParse                = "string to time error"
	ErrIncorrectPasswordOrEmail = "incorrect password or email"
	ErrUserWithEmailNotFound    = "user with this email not found"
	ErrNotFoundInDB             = "not found"
	ErrShortPassword            = "please input password, at least 8 symbols"
)

// http code 401
const (
	ErrTokenExpired     = "token expired"
	ErrNotStandardToken = "token claims are not of type *StandardClaims"
)

// http code 403
const (
	ErrUserIsNotActive = "user is not active. please check your email"
	ErrAccessDenied    = "access denied"
)
