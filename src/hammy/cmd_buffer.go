package hammy

//Command
type Cmd struct {
	CmdType string
	Cmd string
}

//Commads queue from trigger
type CmdBuffer []Cmd

//Create new CmdBuffer
func NewCmdBuffer(size uint32) *CmdBuffer {
	res := make(CmdBuffer, size)
	return &res
}

//Iterface for command commiter
type CmdBufferProcessor interface {
	Process(key string, cmdb *CmdBuffer) error
}
