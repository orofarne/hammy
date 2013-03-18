package hammy

import (
	"fmt"
	"encoding/json"
	"github.com/couchbaselabs/go-couchbase"
	"github.com/dustin/gomemcached"
)

type CouchbaseTriggersGetter struct {
	Client *couchbase.Client
	Pool *couchbase.Pool
	Bucket *couchbase.Bucket
}

func NewCouchbaseTriggersGetter(cfg Config) (*CouchbaseTriggersGetter, error) {
	tg := new(CouchbaseTriggersGetter)

	c, err := couchbase.Connect(cfg.CouchbaseTriggers.ConnectTo)
	if err != nil {
		return nil, err
	}
	tg.Client = &c

	p, err := tg.Client.GetPool(cfg.CouchbaseTriggers.Pool)
	if err != nil {
		return nil, err
	}
	tg.Pool = &p

	b, err := tg.Pool.GetBucket(cfg.CouchbaseTriggers.Bucket)
	if err != nil {
		return nil, err
	}
	tg.Bucket = b

	return tg, nil
}

func (tg *CouchbaseTriggersGetter) MGet(keys []string) (triggers map[string]string, err error) {
	ans := tg.Bucket.GetBulk(keys)

	triggers = make(map[string]string)
	for k, r := range ans {
		switch r.Status {
			case gomemcached.SUCCESS:
				var body string
				err = json.Unmarshal(r.Body, &body)
				if err != nil {
					return
				}
				triggers[k] = body
			case gomemcached.KEY_ENOENT:
				// nil
			default:
				err = fmt.Errorf("%s", r.Error())
				return
		}
	}

	return
}
