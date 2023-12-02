package error

type Error struct {
	position string
	msg      string
	status   int
	error    error
}

func NewError(position string, msg string, status int, err error) *Error {
	return &Error{
		position: position,
		msg:      msg,
		status:   status,
		error:    err,
	}
}
