package hammy

//Interface for trigger executer
type Executer interface {
	//Process trigger for one object
	ProcessTrigger(key string, trigger string, state *State,
		data IncomingObjectData) (newState *State, cmdb *CmdBuffer, err error)
}
