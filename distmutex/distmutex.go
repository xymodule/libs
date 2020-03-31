package distmutex

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/etcd-io/etcd/clientv3/concurrency"
	log "github.com/sirupsen/logrus"
)

const (
	LEASE_TIMEOUT = 5
)

var (
	once            sync.Once
	_default_server server
)

func Init(prefix string, etcd_hosts []string) {
	once.Do(func() {
		_default_server.init(prefix, etcd_hosts)
	})
}

type server struct {
	prefix string
	client *clientv3.Client
	mu     sync.RWMutex
}

func (p *server) init(prefix string, hosts []string) {
	p.prefix = prefix

	cfg := clientv3.Config{
		Endpoints: hosts,
	}

	c, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	p.client = c
	go p.watcher()
}

func (p *server) watcher() {
	watcher := clientv3.NewWatcher(p.client)
	channel := watcher.Watch(context.Background(), p.prefix, clientv3.WithPrefix())
	for {
		select {
		case change := <-channel:
			for _, ev := range change.Events {
				log.Infof("locks change: %s, %v", string(ev.Kv.Key), ev.Type)
			}
		}
	}
}

func (p *server) lockDo(key string, f func()) error {
	if p.client == nil {
		return fmt.Errorf("locks connection err")
	}

	res, err := p.client.Grant(context.Background(), LEASE_TIMEOUT)
	if err != nil {
		return err
	}

	session, err := concurrency.NewSession(p.client, concurrency.WithLease(res.ID))
	if err != nil {
		return err
	}

	mux := concurrency.NewMutex(session, p.prefix+"/"+key)
	if err := mux.Lock(context.Background()); err != nil {
		return err
	}

	session.Orphan()

	//ttl, _ := p.client.TimeToLive(context.TODO(), res.ID)
	//log.Infof("lease ttl = %v", ttl.TTL)

	f()

	if err := mux.Unlock(context.TODO()); err != nil {
		return err
	}

	return nil
}

func DistMutexLockDo(key string, f func()) {
	_default_server.lockDo(key, f)
}
