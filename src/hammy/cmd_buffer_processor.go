package hammy

import "log"

type CmdBufferProcessorImpl struct {
}

func (cbp *CmdBufferProcessorImpl) Process(key string, cmdb *CmdBuffer) (sendBuffer IncomingData, err error) {
	for _, c := range *cmdb {
		switch c.CmdType {
			case "log":
				log.Printf("[%s] %s", key, c.Cmd)
			case "send":
				//TODO
			default:
				log.Printf("[%s] Undefined command: %s\t%s", key, c.CmdType, c.Cmd)
		}
	}

	return NewIncomingData(), nil
}
