package hammy

import "fmt"
import "log"
import "time"

type CmdBufferProcessorImpl struct {
	//Send buffer
	SBuffer SendBuffer
}

func (cbp *CmdBufferProcessorImpl) Process(key string, cmdb *CmdBuffer) error {
	for _, c := range *cmdb {
		switch c.Cmd {
			case "log":
				log.Printf("[%s] %s", key, c.Options["message"])
			case "send":
				cbp.processSend(key, c.Options)
			default:
				log.Printf("[%s] %s %v", key, c.Cmd, c.Options)
		}
	}

	return nil
}

func (cbp *CmdBufferProcessorImpl) log(key string, message string) error {
	cmdb := make(CmdBuffer, 1)
	cmdb[0].Cmd = "log"
	cmdb[0].Options = make(map[string]string)
	cmdb[0].Options["message"] = message
	return cbp.Process(key, &cmdb)
}

func (cbp *CmdBufferProcessorImpl) processSend(key string, opts map[string]string) {
	objName := opts["object"]
	if objName == "" {
		objName = key
	}

	itemKey := opts["key"]
	if itemKey == "" {
		cbp.log(key, fmt.Sprintf("Invalid send: key expected (command options: %v)", opts))
		return
	}

	value, valueFound := opts["value"]
	if !valueFound {
		cbp.log(key, fmt.Sprintf("Invalid send: value expected (command options: %v)", opts))
		return
	}

	data := make(IncomingData)
	objData := make(IncomingObjectData)
	objValue := IncomingValueData{
		Timestamp: uint64(time.Now().Unix()),
		Value: value,
	}
	objData[itemKey] = []IncomingValueData{objValue}
	data[objName] = objData

	cbp.SBuffer.Push(&data)
}
