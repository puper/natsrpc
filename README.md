# natsrpc
nats-server as rpc transport

## server example
```
func New(cfg *Config) (*RpcBroker, error) {
	var err error
	me := &RpcBroker{
		cfg: cfg,
		wg:  new(sync.WaitGroup),
	}
	me.sub, err = cfg.NatsCli.QueueSubscribe(cfg.Subject, cfg.Queue, func(msg *nats.Msg) {
		me.wg.Add(1)
		defer me.wg.Done()
		logCtx := me.cfg.Logger.With("msg.data", json.RawMessage(msg.Data))
		req := new(protos.Request)
		if err := json.Unmarshal(msg.Data, req); err != nil {
			logCtx.Errorw("request.Unmarshal.Error", "error", err)
            resp.Error = protos.NewRpcError(err)
            msg.Respond(resp.Encode())
			return
		}
		resp := new(protos.Response)
        // services.GetCaller() return a instance of https://github.com/puper/servicecaller
        // service may return a protos.RpcError
		err := services.GetCaller().Call(req.ServiceMethod, req.Args, &resp.Result)
		if err != nil {
			logCtx.Errorw("request.Call.Error", "error", err)
			resp.Error = protos.NewRpcError(err)
		} else {
			logCtx.Debugw("request.success", "result", resp.Result)
		}
		msg.Respond(resp.Encode())
	})
	if err != nil {
		return nil, err
	}
	return me, nil
}

type RpcBroker struct {
	cfg *Config
	sub *nats.Subscription
	wg  *sync.WaitGroup
}

func (me *RpcBroker) Close() error {
	err := me.sub.Drain()
	me.wg.Wait()
	return err
}
```