package hammy

import (
	"fmt"
	"log"
	"time"
)

type CmdBufferProcessorImpl struct {
	// Send buffer
	SBuffer SendBuffer
	// Data saver
	Saver SendBuffer
}

func (cbp *CmdBufferProcessorImpl) Process(key string, cmdb *CmdBuffer) error {
	for _, c := range *cmdb {
		switch c.Cmd {
			case "log":
				log.Printf("[%s] %s", key, c.Options["message"])
			case "send":
				cbp.processSend(key, c.Options)
			case "save":
				cbp.processSave(key, c.Options)
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

func (cbp *CmdBufferProcessorImpl) processSave(key string, opts map[string]string) {
	itemKey := opts["key"]
	if itemKey == "" {
		cbp.log(key, fmt.Sprintf("Invalid save: key expected (command options: %v)", opts))
		return
	}

	value, valueFound := opts["value"]
	if !valueFound {
		cbp.log(key, fmt.Sprintf("Invalid save: value expected (command options: %v)", opts))
		return
	}

	var ts uint64
	ts_s := opts["timestamp"]
	if ts_s == "" {
		ts = uint64(time.Now().Unix())
	} else {
		_, err := fmt.Sscan(ts_s, &ts)
		if err != nil {
			cbp.log(key, fmt.Sprintf("Invalid save: invalid timestamp (command options: %v)", opts))
			return
		}
	}

	data := make(IncomingData)
	objData := make(IncomingObjectData)
	objValue := IncomingValueData{
		Timestamp: ts,
		Value: value,
	}
	objData[itemKey] = []IncomingValueData{objValue}
	data[key] = objData

	cbp.Saver.Push(&data)
}
