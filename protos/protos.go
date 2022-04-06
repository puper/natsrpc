package protos

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	NatsRpcErrorKey = "NatsRpcError:"
)

var (
	ErrNotRpcError = errors.New("not rpc error")
)

type Request struct {
	ServiceMethod string          `json:"serviceMethod"`
	Args          json.RawMessage `json:"args"`
}

func (me *Request) Encode() []byte {
	b, _ := json.Marshal(me)
	return b
}

type Response struct {
	Error  *RpcError       `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

func (me *Response) Encode() []byte {
	b, _ := json.Marshal(me)
	return b
}

func ParseRpcError() *RpcError {
	return nil
}

type CodeMessageError interface {
	GetCode() string
	GetMessage() string
}

type DetailsError interface {
	CodeMessageError
	GetDetails() any
}

func FromError(err error) *RpcError {
	return nil
}

type rpcError struct {
	Code    string          `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
	Details json.RawMessage `json:"details,omitempty"`
}

type RpcError struct {
	rpcError
}

func NewRpcError(opts ...any) *RpcError {
	me := &RpcError{}
	l := len(opts)
	if l > 0 {
		if err, ok := opts[0].(*RpcError); ok {
			*me = *err
		} else if err, ok := opts[0].(DetailsError); ok {
			me.Code = err.GetCode()
			me.Message = err.GetMessage()
			b, _ := json.Marshal(err.GetDetails())
			me.Details = b
		} else if err, ok := opts[0].(CodeMessageError); ok {
			me.Code = err.GetCode()
			me.Message = err.GetMessage()
		} else if err, ok := opts[0].(error); ok {
			me.Message = err.Error()
		} else {
			me.Code = opts[0].(string)
		}
	}
	if l > 1 {
		me.Message = opts[1].(string)
	}
	if l > 2 {
		b, _ := json.Marshal(opts[2])
		me.Details = b
	}
	return me
}

func (me *RpcError) GetCode() string {
	return me.Code
}

func (me *RpcError) GetMessage() string {
	return me.Message
}

func (me *RpcError) GetDetails() any {
	return me.Details
}

func (me *RpcError) Error() string {
	v := map[string]any{
		NatsRpcErrorKey: me,
	}
	b, _ := json.Marshal(v)
	return string(b)

}

func (me *RpcError) MarshalJSON() ([]byte, error) {
	return json.Marshal(me.rpcError)
}

func (me *RpcError) UnmarshalJSON(b []byte) error {
	v := map[string]rpcError{}
	if err := json.Unmarshal(b, &v); err == nil {
		if _, ok := v[NatsRpcErrorKey]; ok {
			me.rpcError = v[NatsRpcErrorKey]
			return nil
		}
	}
	return ErrNotRpcError
}
