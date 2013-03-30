package hammy

import (
	"log"
	"time"
	"container/list"
	"encoding/json"
)

// Buffer for reprocessed data
type SendBufferImpl struct {
	dataChan chan *IncomingData
	data *list.List
	// Timeout between sends
	sleepTime time.Duration
	rHandler RequestHandler

	//Metics
	ms *MetricSet
	mPushes *CounterMetric
	mSendedValues *CounterMetric
	mSend *TimerMetric
	mErrors *CounterMetric
}

// Creates and initialize new SendBuffer
func NewSendBufferImpl(rh RequestHandler, cfg Config, metricsNamespace string) (sb *SendBufferImpl) {
	sb = new(SendBufferImpl)
	sb.dataChan = make(chan *IncomingData)
	sb.data = list.New()
	sb.sleepTime = time.Duration(1000 * cfg.SendBuffer.SleepTime) * time.Millisecond
	sb.rHandler = rh

	sb.ms = NewMetricSet(metricsNamespace, 30*time.Second)
	sb.mPushes = sb.ms.NewCounter("pushes")
	sb.mSendedValues = sb.ms.NewCounter("sended_values")
	sb.mSend = sb.ms.NewTimer("send")
	sb.mErrors = sb.ms.NewCounter("errors")

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
	sb.mPushes.Add(1)
}

// Process detached data buffer
func (sb *SendBufferImpl) send(data *list.List) {
	if data.Len() == 0 {
		return
	}

	// Statistics
	τ := sb.mSend.NewObservation()
	defer func() { τ.End() } ()
	var sended_values uint64
	defer func() { sb.mSendedValues.Add(sended_values) } ()

	// 1) Merge list
	mData := make(IncomingData)
	// iterates over data list
	for e := data.Front(); e != nil; e = e.Next() {
		// e.Value is *IncomingData (panic otherwise)
		eData := e.Value.(*IncomingData)
		for hostK, hostV := range *eData {
			mV, hostFound := mData[hostK]
			if hostFound {
				for k, v := range hostV {
					iArr, kFound := mV[k]
					if kFound {
						newIArr := make([]IncomingValueData, len(v) + len(iArr))
						for i, vE := range iArr {
							newIArr[i] = vE
							sended_values++ // Statistics
						}
						for j, vE := range v {
							newIArr[len(iArr) + j] = vE
							sended_values++ // Statistics
						}
						mV[k] = newIArr
					} else {
						mV[k] = v
					}
				}
			} else {
				mData[hostK] = hostV
			}
		}
	}

	// 2) Process merged data
	errs := sb.rHandler.Handle(mData)

	if len(errs) > 0 {
		sb.mErrors.Add(uint64(len(errs))) // Statistics

		// FIXME
		errs_str := make(map[string]string)
		for k, v := range errs {
			errs_str[k] = v.Error()
		}
		b, err := json.Marshal(errs_str)
		if err == nil {
			log.Printf("!! Error in SendBuffer: %s", b)
		} else {
			log.Printf("!! Error in SendBuffer: <unable to dump errs: %#v>", err)
		}
	}
}
