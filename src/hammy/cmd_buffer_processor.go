package hammy

import (
	"io"
	"fmt"
	"log"
	"time"
	"strings"
	"reflect"
	"encoding/json"
)

type CmdBufferProcessorImpl struct {
	// Send buffer
	SBuffer SendBuffer
	// Data saver
	Saver SendBuffer
	// FIXME Other commands
	CmdOutput io.Writer
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
				b, err := json.Marshal(c)
				if err == nil {
					b = append(b, '\n')
					_, err = cbp.CmdOutput.Write(b)
				}
				if err != nil {
					cbp.log(key, fmt.Sprintf("CmdOutput error: %v", err))
				}
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
		case int, int8, int16, int32, int64:
			value = float64(reflect.ValueOf(value).Int())
		case uint, uint8, uint16, uint32, uint64:
			value = float64(reflect.ValueOf(value).Uint())
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
		case uint, uint8, uint16, uint32, uint64:
			ts = reflect.ValueOf(tsRaw).Uint()
		case int, int8, int16, int32, int64:
			ts = uint64(reflect.ValueOf(tsRaw).Int())
		case float32, float64:
			ts = uint64(reflect.ValueOf(tsRaw).Float())
		case string:
			_, err := fmt.Sscan(tsRaw.(string), &ts)
			if err != nil {
				cbp.log(key, fmt.Sprintf("Invalid save: invalid timestamp (command options: %v)", opts))
				return
			}
		default:
			cbp.log(key, fmt.Sprintf("Invalid save: invalid timestamp of type %T (command options: %v)", tsRaw, opts))
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
