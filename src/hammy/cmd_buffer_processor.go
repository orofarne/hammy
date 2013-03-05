package hammy

import "log"

type CmdBufferProcessorImpl struct {
}

func (cbp *CmdBufferProcessorImpl) Process(key string, cmdb *CmdBuffer) error {
	for _, c := range *cmdb {
		log.Printf("[%s] %s\t\t%s", key, c.CmdType, c.Cmd)
	}

	return nil
}
