package hammy

// Interface for trigger executer
type Executer interface {
	// Process trigger for one host
	ProcessTrigger(key string, trigger string, state *State,
		data IncomingHostData) (newState *State, cmdb *CmdBuffer, err error)
}
