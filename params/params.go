package params
import (
	"gopkg.in/juju/charm.v2"
)

type Error struct {
	Message string
	Code    string
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) ErrorCode() string {
	return e.Code
}

type ErrorCoder interface {
	ErrorCode() string
}

type MetaAnyResponse struct {
	Id *charm.URL
	Meta map[string] interface{}
}
