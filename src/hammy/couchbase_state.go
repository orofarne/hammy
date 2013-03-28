package hammy

import (
	"fmt"
	"encoding/json"
	"encoding/binary"
	"github.com/couchbaselabs/go-couchbase"
	"github.com/dustin/gomemcached"
	"github.com/dustin/gomemcached/client"
)

type CouchbaseStateKeeper struct {
	Client *couchbase.Client
	Pool *couchbase.Pool
	Bucket *couchbase.Bucket
	Ttl uint32
}

func NewCouchbaseStateKeeper(cfg Config) (*CouchbaseStateKeeper, error) {
	sk := new(CouchbaseStateKeeper)

	c, err := couchbase.Connect(cfg.CouchbaseStates.ConnectTo)
	if err != nil {
		return nil, err
	}
	sk.Client = &c

	p, err := sk.Client.GetPool(cfg.CouchbaseStates.Pool)
	if err != nil {
		return nil, err
	}
	sk.Pool = &p

	b, err := sk.Pool.GetBucket(cfg.CouchbaseStates.Bucket)
	if err != nil {
		return nil, err
	}
	sk.Bucket = b

	sk.Ttl = uint32(cfg.CouchbaseStates.Ttl)

	return sk, nil
}

func (sk *CouchbaseStateKeeper) Get(key string) StateKeeperAnswer {
	s := NewState()
	var cas uint64
	err := sk.Bucket.Gets(key, s, &cas)

	if err == nil {
		return StateKeeperAnswer{
			State: *s,
			Cas: &cas,
			Err: nil,
		}
	} else {
		return StateKeeperAnswer{
			State: nil,
			Cas: nil,
			Err: err,
		}
	}
	panic("?!!")
}

func (sk *CouchbaseStateKeeper) MGet(keys []string) (states map[string]StateKeeperAnswer) {
	ans := sk.Bucket.GetBulk(keys)

	states = make(map[string]StateKeeperAnswer)
	for k, r := range ans {
		switch r.Status {
			case gomemcached.SUCCESS:
				s := NewState()
				err := json.Unmarshal(r.Body, s)
				if err == nil {
					states[k] = StateKeeperAnswer{
						State: *s,
						Cas: &r.Cas,
						Err: nil,
					}
				} else {
					states[k] = StateKeeperAnswer{
						State: nil,
						Cas: nil,
						Err: err,
					}
				}
			case gomemcached.KEY_ENOENT:
				states[k] = StateKeeperAnswer{
					State: *NewState(),
					Cas: nil,
					Err: nil,
				}
			default:
				states[k] = StateKeeperAnswer{
					State: nil,
					Cas: nil,
					Err: fmt.Errorf("%s", r.Error()),
				}
		}
	}

	for _, k := range keys {
		if _, found := states[k]; !found {
			states[k] = StateKeeperAnswer{
				State: *NewState(),
				Cas: nil,
				Err: nil,
			}
		}
	}

	return
}

func (sk *CouchbaseStateKeeper) Set(key string, data State, cas *uint64) (retry bool, err error) {
	err = sk.Bucket.Do(key, func(mc *memcached.Client, vb uint16) (e error) {
		buf, e := json.Marshal(data)
		if e != nil {
			return
		}
		req := &gomemcached.MCRequest{
			Opcode: gomemcached.SET,
			VBucket: vb,
			Key: []byte(key),
			Cas: 0,
			Opaque: 0,
			Extras: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			Body: buf,
		}
		if cas != nil {
			req.Cas = *cas
		}
		binary.BigEndian.PutUint32(req.Extras[4:8], sk.Ttl)

		resp, e := mc.Send(req)
		if e != nil {
			if resp != nil && resp.Status == gomemcached.KEY_EEXISTS {
				e = nil
				retry = true
			}
			return
		}

		switch resp.Status {
			case gomemcached.KEY_EEXISTS:
				retry = true
				return
			case gomemcached.SUCCESS:
				return
			default:
				return fmt.Errorf("CAS operation failed: %v", resp.Error())
		}
		panic("?!!")
	})
	if err != nil {
		return
	}

	return
}
