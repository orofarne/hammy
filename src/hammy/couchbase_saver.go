package hammy

import (
	"fmt"
	"log"
	"github.com/ugorji/go-msgpack"
	"github.com/couchbaselabs/go-couchbase"
	"github.com/dustin/gomemcached"
	"github.com/dustin/gomemcached/client"
)

// Saves historical data to write chache (based on couchbase)
type CouchbaseSaver struct {
	client *couchbase.Client
	pool *couchbase.Pool
	bucket *couchbase.Bucket
	dataChan chan *IncomingData
}

// Create new saver
func NewCouchbaseSaver(cfg Config) (*CouchbaseSaver, error) {
	s := new(CouchbaseSaver)

	c, err := couchbase.Connect(cfg.CouchbaseSaver.ConnectTo)
	if err != nil {
		return nil, err
	}
	s.client = &c

	p, err := s.client.GetPool(cfg.CouchbaseSaver.Pool)
	if err != nil {
		return nil, err
	}
	s.pool = &p

	b, err := s.pool.GetBucket(cfg.CouchbaseSaver.Bucket)
	if err != nil {
		return nil, err
	}
	s.bucket = b

	// Process queue
	s.dataChan = make(chan *IncomingData, cfg.CouchbaseSaver.QueueSize)
	// Workers...
	for i := uint(0); i < cfg.CouchbaseSaver.SavePoolSize; i++ {
		go s.worker()
	}

	return s, nil
}

// Enqueue data for saving
func (s *CouchbaseSaver) Push(data *IncomingData) {
	s.dataChan <- data
}

const CouchbaseDataBucketQuantum = 7200 // 2 hours

func CouchbaseSaverBucketKey(hostKey string, itemKey string, timestamp uint64) string {
	var bucketId uint64
	bucketId = timestamp / CouchbaseDataBucketQuantum
	return fmt.Sprintf("%s$%s$%d", hostKey, itemKey, bucketId)
}

func (s *CouchbaseSaver) worker() {
	for data := range s.dataChan {
		for hostK, hostV := range *data {
			for itemK, itemV := range hostV {
				for _, v := range itemV {
					err := s.saveItem(hostK, itemK, v)
					if err != nil {
						log.Printf("saveItem error: %v", err)
					}
				}
			}
		}
	}
}

func (s *CouchbaseSaver) saveItem(hostKey string, itemKey string, val IncomingValueData) error {
	bucketKey := CouchbaseSaverBucketKey(hostKey, itemKey, val.Timestamp)

	buf, err := msgpack.Marshal(val)
	if err != nil {
		return err
	}

	err = s.bucket.Do(bucketKey, func(mc *memcached.Client, vb uint16) (e error) {
RETRY:
		req := &gomemcached.MCRequest{
			Opcode: gomemcached.APPEND,
			VBucket: vb,
			Key: []byte(bucketKey),
			Cas: 0,
			Opaque: 0,
			Extras: nil,
			Body: buf,
		}

		resp, e := mc.Send(req)
		if resp == nil {
			return
		}

		switch resp.Status {
			case gomemcached.SUCCESS:
				return
			case gomemcached.NOT_STORED:
				req := &gomemcached.MCRequest{
					Opcode: gomemcached.ADD,
					VBucket: vb,
					Key: []byte(bucketKey),
					Cas: 0,
					Opaque: 0,
					Extras: []byte{0, 0, 0, 0, 0, 0, 0, 0},
					Body: buf,
				}

				resp, e = mc.Send(req)
				if resp == nil {
					return
				}

				switch resp.Status {
					case gomemcached.SUCCESS:
						return
					case gomemcached.KEY_EEXISTS:
						goto RETRY
					default:
						return fmt.Errorf("ADD operation failed: %v", resp.Error())
				}
			default:
				return fmt.Errorf("APPEND operation failed: %v", resp.Error())
		}
		panic("?!!")
	});

	return err
}
