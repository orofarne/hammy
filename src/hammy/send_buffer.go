package hammy

import (
	"log"
	"time"
	"container/list"
)

// Buffer for reprocessed data
type SendBufferImpl struct {
	dataChan chan *IncomingData
	data *list.List
	// Timeout between sends
	sleepTime time.Duration
	rHandler RequestHandler
}

// Creates and initialize new SendBuffer
func NewSendBufferImpl(rh RequestHandler, cfg Config) (sb *SendBufferImpl) {
	sb = new(SendBufferImpl)
	sb.dataChan = make(chan *IncomingData)
	sb.data = list.New()
	sb.sleepTime = time.Duration(1000 * cfg.SendBuffer.SleepTime) * time.Millisecond
	sb.rHandler = rh
	return
}

// Locks and begins data processing
func (sb *SendBufferImpl) Listen() {
	timer := time.Tick(sb.sleepTime)

	for {
		select {
			case newData := <- sb.dataChan:
				sb.data.PushBack(newData)
			case <- timer:
				go sb.send(sb.data)
				sb.data = list.New()
		}
	}
}

// Enqueue data for reprocessing
func (sb *SendBufferImpl) Push(data *IncomingData) {
	sb.dataChan <- data
}

// Process detached data buffer
func (sb *SendBufferImpl) send(data *list.List) {
	if data.Len() == 0 {
		return
	}

	// 1) Merge list
	mData := make(IncomingData)
	// iterates over data list
	for e := data.Front(); e != nil; e = e.Next() {
		// e.Value is *IncomingData (panic otherwise)
		eData := e.Value.(*IncomingData)
		for objK, objV := range *eData {
			mV, objFound := mData[objK]
			if objFound {
				for k, v := range objV {
					iArr, kFound := mV[k]
					if kFound {
						newIArr := make([]IncomingValueData, len(v) + len(iArr))
						for i, vE := range iArr {
							newIArr[i] = vE
						}
						for j, vE := range v {
							newIArr[len(iArr) + j] = vE
						}
						mV[k] = newIArr
					} else {
						mV[k] = v
					}
				}
			} else {
				mData[objK] = objV
			}
		}
	}

	// 2) Process merged data
	errs := sb.rHandler.Handle(mData)

	if len(errs) > 0 {
		log.Printf("!! Error in SendBuffer: %v", errs) // FIXME
	}
}
