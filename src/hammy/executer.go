package hammy

//Interface for trigger executer
type Executer interface {
	ProcessTrigger(key string, trigger string, state *State,
		data IncomingObjectData) (cmdb *CmdBuffer, err error)
}
