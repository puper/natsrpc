package natsrpc

import (
	"time"

	"github.com/nats-io/nats.go"
)

type Config struct {
	NatsCli       *nats.Conn    `json:"-"`
	StreamSubject string        `json:"streamSubject"`
	RpcSubject    string        `json:"rpcSubject"`
	RpcTimeout    time.Duration `json:"rpcTimeout"`
}
