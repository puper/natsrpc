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
	defaultSubject := fmt.Sprintf("%v.default.*", me.cfg.Subject)
	workerSubject := fmt.Sprintf("%v.worker.%v", me.cfg.Subject, me.cfg.DispatchWorkerId)
	me.defaultSub, err = cfg.NatsCli.QueueSubscribe(defaultSubject, cfg.Queue, func(msg *nats.Msg) {
		me.wg.Add(1)
		defer me.wg.Done()
		if cfg.DispatchWorkerNum > 0 && msg.Subject != fmt.Sprintf("%v.default._default_", me.cfg.Subject) {
			workId := crc32.ChecksumIEEE([]byte(msg.Subject)) % cfg.DispatchWorkerNum
			if workId != me.cfg.DispatchWorkerId {
				msg.Subject = fmt.Sprintf("%v.worker.%v", me.cfg.Subject, workId)
				cfg.NatsCli.PublishMsg(msg)
				return
			}
		}
		me.handleMsg(msg)
	})
	if err != nil {
		return nil, err
	}
	if me.cfg.DispatchWorkerNum > 0 {
		// 确认是自己的消息，直接处理
		me.workerSub, err = cfg.NatsCli.Subscribe(workerSubject, func(msg *nats.Msg) {
			me.wg.Add(1)
			defer me.wg.Done()
			me.handleMsg(msg)
		})
		if err != nil {
			return nil, err
		}
	}
	return me, nil
}

func (me *RpcBroker) handleMsg(msg *nats.Msg) {
	resp := new(protos.Response)
	if atomic.LoadInt64(&me.closed) == 1 {
		resp.Error = protos.NewRpcError(fmt.Errorf("server.Closed"))
		msg.Respond(resp.Encode())
		return
	}
	logCtx := me.cfg.Logger.With("msg.data", json.RawMessage(msg.Data))
	req := new(protos.Request)
	if err := json.Unmarshal(msg.Data, req); err != nil {
		logCtx.Errorw("request.Unmarshal.Error", "error", err)
		resp.Error = protos.NewRpcError(err)
		msg.Respond(resp.Encode())
		return
	}
	err := services.GetCaller().Call(req.ServiceMethod, req.Args, &resp.Result)
	if err != nil {
		logCtx.Errorw("request.Call.Error", "error", err)
		resp.Error = protos.NewRpcError(err)
		msg.Respond(resp.Encode())
		return
	}
	logCtx.Debugw("request.success", "result", resp.Result)
	msg.Respond(resp.Encode())
}
```