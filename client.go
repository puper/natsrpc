package natsrpc

import (
	"encoding/json"
	"fmt"

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
	cfg     *Config
	natsCli *nats.Conn
	js      nats.JetStream
}

func (me *Client) Stream(serviceMethod string, args any) error {
	b, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req := &protos.Request{
		ServiceMethod: serviceMethod,
		Args:          b,
	}
	_, err = me.js.Publish(me.cfg.StreamSubject, req.Encode())
	if err != nil {
		return err
	}
	return nil
}

func (me *Client) Call(serviceMethod string, args any, reply any) *protos.RpcError {
	b, _ := json.Marshal(args)
	req := &protos.Request{
		ServiceMethod: serviceMethod,
		Args:          b,
	}
	msg, err := me.natsCli.Request(me.cfg.RpcSubject, req.Encode(), me.cfg.RpcTimeout)
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
