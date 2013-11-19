package hammy

type Metric map[string]string

type Data struct {
	M Metric
	D interface{}
}

type DataBus interface {
	Push(d *Data)
	Pull() []*Data
}

type StateKeeper interface {
	Set(m Metric, data *interface{}) bool
	Get(m Metric) *interface{}
}

type CodeLoader interface {
	Get(m Metric) []string
}

type HammyApp struct {
	DB DataBus
	SK StateKeeper
	CL CodeLoader
}
