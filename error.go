package sorm

type Error struct {
	err error
	msg string
}

func (e Error) Error() string {
	return e.msg
}

func (e Error) Unwrap() error {
	return e.err
}

func NewError(err error, msg string) Error {
	return Error{
		err: err,
		msg: msg,
	}
}
