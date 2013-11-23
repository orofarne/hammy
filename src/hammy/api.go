package hammy

import (
	"time"
)

type Metric struct {
	Host string
	Item string
}

type Data struct {
	Metric
	Timestamp time.Time
	Value     interface{}
}

type State interface{}

type Dependence struct {
	Item    string
	Timeout time.Duration
}

type Trigger struct {
	Code         string
	Dependencies []Dependence
}

type HostState map[string]State
type HostTrigger map[string]Trigger

type DataBus interface {
	Push(d []Data)
	Pull() []Data
}

type StateKeeper interface {
	Set(host string, st HostState) bool
	Get(host string) HostState
}

type CodeLoader interface {
	Get(host string) HostTrigger
}

type HammyApp struct {
	DB DataBus
	SK StateKeeper
	CL CodeLoader
}

func NewHostState() HostState {
	return make(HostState)
}

func NewHostTrigger() HostTrigger {
	return make(HostTrigger)
}
