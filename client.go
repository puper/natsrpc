package natsrpc

import (
	"encoding/json"
	"fmt"
	"hash/crc32"

	"github.com/nats-io/nats.go"
	"github.com/puper/natsrpc/protos"
)

func New(cfg *Config) (*Client, error) {
	me := &Client{
		cfg: cfg,
	}
	js, err := cfg.NatsCli.JetStream()
	if err != nil {
		return nil, err
	}
	me.js = js
	return me, nil
}

type Client struct {
	cfg *Config
	js  nats.JetStream
}

/*
provide dispatch key if you want dispatch task.
*/
func (me *Client) Stream(serviceMethod string, args any, dispatchKeys ...string) error {
	b, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req := &protos.Request{
		ServiceMethod: serviceMethod,
		Args:          b,
	}
	subject := me.cfg.StreamSubject + ".default.__default__"
	if len(dispatchKeys) > 0 {
		subject += "." + dispatchKeys[0]
	}
	_, err = me.js.Publish(subject, req.Encode())
	if err != nil {
		return err
	}
	return nil
}

func (me *Client) Call(serviceMethod string, args any, reply any, dispatchKeys ...string) *protos.RpcError {
	b, _ := json.Marshal(args)
	req := &protos.Request{
		ServiceMethod: serviceMethod,
		Args:          b,
	}
	subject := me.cfg.RpcSubject + ".default.__default__"
	if len(dispatchKeys) > 0 {
		subject += "." + dispatchKeys[0]
	}
	msg, err := me.cfg.NatsCli.Request(subject, req.Encode(), me.cfg.RpcTimeout)
	if err != nil {
		return protos.NewRpcError("", fmt.Sprintf("RpcRequestError: %v", err.Error()))
	}
	resp := new(protos.Response)
	err = json.Unmarshal(msg.Data, resp)
	if err != nil {
		return protos.NewRpcError("", fmt.Sprintf("RpcResponseUnmarshalError: %v", err.Error()))
	}
	if resp.Error != nil {
		return resp.Error
	}
	err = json.Unmarshal(resp.Result, reply)
	if err != nil {
		return protos.NewRpcError("", fmt.Sprintf("RpcResponseUnmarshalError: %v", err.Error()))
	}
	return nil
}

func DispatchKey(format string, a ...any) string {
	return fmt.Sprint(crc32.ChecksumIEEE([]byte(fmt.Sprintf(format, a...))))
}
