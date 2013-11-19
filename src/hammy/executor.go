package hammy

type execMessage struct {
}

type Executor interface {
	Process(trigger string, data *Data, state *interface{}) (*Data, *interface{})
}

type ExecutorImpl struct {
}

func (e *ExecutorImpl) Process(trigger string, data *Data, state *interface{}) (*Data, *interface{}) {
	// TODO
	return nil, state
}
