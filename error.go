package sorm

import "errors"

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

func TryCatch(try func(), catch func(err error), errs ...error) {
	defer func() {
		if recv := recover(); recv != nil {
			if e, ok := recv.(error); ok {
				for _, err := range errs {
					if errors.Is(e, err) {
						catch(e)
						return
					}
				}
			}
			panic(recv)
		}
	}()
	try()
}
