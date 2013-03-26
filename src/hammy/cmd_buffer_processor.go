package hammy

import (
	"fmt"
	"log"
	"time"
	"strings"
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
	cmdb[0].Options = make(map[string]interface{})
	cmdb[0].Options["message"] = message
	return cbp.Process(key, &cmdb)
}

func (cbp *CmdBufferProcessorImpl) processSend(key string, opts map[string]interface{}) {
	hostNameRaw := opts["host"]
	var hostName string
	if hostNameRaw == nil {
		hostName = key
	} else {
		var converted bool
		hostName, converted = hostNameRaw.(string)
		if !converted {
			cbp.log(key, fmt.Sprintf("Invalid send: invalid host name (command options: %v)", opts))
			return
		}
	}

	itemKey, itemKeyConverted := opts["key"].(string)
	if !itemKeyConverted  || itemKey == "" {
		cbp.log(key, fmt.Sprintf("Invalid send: key expected (command options: %v)", opts))
		return
	}

	value, valueFound := opts["value"]
	if !valueFound {
		cbp.log(key, fmt.Sprintf("Invalid send: value expected (command options: %v)", opts))
		return
	}

	data := make(IncomingData)
	hostData := make(IncomingHostData)
	hostValue := IncomingValueData{
		Timestamp: uint64(time.Now().Unix()),
		Value: value,
	}
	hostData[itemKey] = []IncomingValueData{hostValue}
	data[hostName] = hostData

	cbp.SBuffer.Push(&data)
}

func (cbp *CmdBufferProcessorImpl) processSave(key string, opts map[string]interface{}) {
	itemKey, itemKeyConverted := opts["key"].(string)
	if !itemKeyConverted || itemKey == "" {
		cbp.log(key, fmt.Sprintf("Invalid save: key expected (command options: %v)", opts))
		return
	}

	value, valueFound := opts["value"]
	if !valueFound {
		cbp.log(key, fmt.Sprintf("Invalid save: value expected (command options: %v)", opts))
		return
	}

	switch value.(type) {
		case int:
			value = float64(value.(int))
		case int8:
			value = float64(value.(int8))
		case int16:
			value = float64(value.(int16))
		case int32:
			value = float64(value.(int32))
		case int64:
			value = float64(value.(int64))
		case uint:
			value = float64(value.(uint))
		case uint8:
			value = float64(value.(uint8))
		case uint16:
			value = float64(value.(uint16))
		case uint32:
			value = float64(value.(uint32))
		case uint64:
			value = float64(value.(uint64))
		case float32:
			value = float64(value.(float32))
		case float64:
			// Do nothing
		case string:
			// Do nothing
		default:
			value = fmt.Sprint(value)
	}

	if strings.HasSuffix(itemKey, "#log") {
		if _, converted := value.(string); !converted {
			value = fmt.Sprint(value)
		}
	} else {
		if _, converted := value.(float64); !converted {
			var val float64
			str, strConverted := value.(string)
			if !strConverted {
				cbp.log(key, fmt.Sprintf("Invalid save: invalid value for non log key `%s` (command options: %v)", itemKey, opts))
				return
			}
			if n, _ := fmt.Sscan(str, &val); n != 1 {
				cbp.log(key, fmt.Sprintf("Invalid save: invalid value for non log key `%s` (command options: %v)", itemKey, opts))
				return
			}
			value = val
		}
	}

	var ts uint64
	tsRaw := opts["timestamp"]
	switch tsRaw.(type) {
		case nil:
			ts = uint64(time.Now().Unix())
		case string:
			_, err := fmt.Sscan(tsRaw.(string), &ts)
			if err != nil {
				cbp.log(key, fmt.Sprintf("Invalid save: invalid timestamp (command options: %v)", opts))
				return
			}
		default:
			cbp.log(key, fmt.Sprintf("Invalid save: invalid timestamp (command options: %v)", opts))
			return
	}

	data := make(IncomingData)
	hostData := make(IncomingHostData)
	hostValue := IncomingValueData{
		Timestamp: ts,
		Value: value,
	}
	hostData[itemKey] = []IncomingValueData{hostValue}
	data[key] = hostData

	cbp.Saver.Push(&data)
}
