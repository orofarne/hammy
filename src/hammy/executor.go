package hammy

type execMessage struct {
}

type Executor interface {
	Process(trigger string, data Data, state State) (Data, State)
}

type ExecutorImpl struct {
}

func (e *ExecutorImpl) Process(trigger string, data Data, state State) (Data, State) {
	// TODO
	return Data{}, state
}
