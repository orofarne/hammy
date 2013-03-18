package hammy

import (
	"fmt"
	"sort"
	"github.com/couchbaselabs/go-couchbase"
	"github.com/ugorji/go-msgpack"
	"github.com/dustin/gomemcached"
)

//Reads data from write cache or storage
type DataReader interface {
	Read(objKey string, itemKey string, from uint64, to uint64) (data []IncomingValueData, err error)
}

//Reads data from write cache (couchbase-based)
type CouchbaseDataReader struct {
	client *couchbase.Client
	pool *couchbase.Pool
	bucket *couchbase.Bucket
}

//Create new saver
func NewCouchbaseDataReader(cfg Config) (*CouchbaseDataReader, error) {
	s := new(CouchbaseDataReader)

	c, err := couchbase.Connect(cfg.CouchbaseDataReader.ConnectTo)
	if err != nil {
		return nil, err
	}
	s.client = &c

	p, err := s.client.GetPool(cfg.CouchbaseDataReader.Pool)
	if err != nil {
		return nil, err
	}
	s.pool = &p

	b, err := s.pool.GetBucket(cfg.CouchbaseDataReader.Bucket)
	if err != nil {
		return nil, err
	}
	s.bucket = b

	return s, nil
}

type dataTimeSorter struct {
	Data *[]IncomingValueData
}

func (ds *dataTimeSorter) Len() int {
	return len(*ds.Data)
}

func (ds *dataTimeSorter) Less(i, j int) bool {
	return (*ds.Data)[i].Timestamp < (*ds.Data)[j].Timestamp
}

func (ds *dataTimeSorter) Swap(i, j int) {
	(*ds.Data)[i], (*ds.Data)[j] = (*ds.Data)[j], (*ds.Data)[i]
}

func (cr *CouchbaseDataReader) Read(objKey string, itemKey string, from uint64, to uint64) (data []IncomingValueData, err error) {
	//Construct keys slice
	bucketFrom, bucketTo := (from / CouchbaseDataBucketQuantum), (to / CouchbaseDataBucketQuantum)
	keys := make([]string, (bucketTo - bucketFrom + 1))
	for i, k := 0, bucketFrom; k <= bucketTo; k++ {
		keys[i] = fmt.Sprintf("%s$%s$%d", objKey, itemKey, k)
		i++
	}

	//Retrive data
	ans := cr.bucket.GetBulk(keys)
	dataLen := 0

	for _, r := range ans {
		switch r.Status {
			case gomemcached.SUCCESS:
				dataLen += len(r.Body)
			case gomemcached.KEY_ENOENT:
				//nil
			default:
				err = fmt.Errorf("%s", r.Error())
				return
		}
	}

	dataRaw := make([]byte, dataLen)
	j := 0
	for _, r := range ans {
		if r.Status == gomemcached.SUCCESS {
			for i := 0; i < len(r.Body); i++ {
				dataRaw[j] = r.Body[i]
				j++
				if j > dataLen {
					panic("Invalid j")
				}
			}
		}
	}

	err = msgpack.Unmarshal(dataRaw, &data, nil)
	if err != nil {
		return
	}

	ds := dataTimeSorter{
		Data: &data,
	}
	sort.Sort(&ds)

	return
}
