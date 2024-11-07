package utils

type Status int

type ResponseError struct {
	Code    Status
	Message string
}

func (re ResponseError) Error() string {
	return re.Message
}
