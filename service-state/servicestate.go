package servicestate

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	etcdclient "github.com/coreos/etcd/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var (
	once            sync.Once
	_default_server server
)

func Init(root_path, service_id string, etcd_hosts []string) {
	once.Do(func() {
		_default_server.init(root_path, service_id, etcd_hosts)
	})
}

type server struct {
	root       string
	service_id string
	client     etcdclient.Client
	datas      map[string]map[string]int
	mu         sync.RWMutex
}

func (p *server) init(root_path, service_id string, etcd_hosts []string) {
	p.root = root_path
	p.service_id = service_id

	p.datas = make(map[string]map[string]int)

	cfg := etcdclient.Config{
		Endpoints: etcd_hosts,
		Transport: etcdclient.DefaultTransport,
	}

	c, err := etcdclient.New(cfg)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}

	p.client = c

	//
	p.load()
}

func (p *server) path(key string) (category, service string, err error) {
	params := strings.Split(key, "/")
	if len(params) != 4 {
		err = fmt.Errorf("Split %v len not equal 4", key)
		return
	}

	category = params[2]
	service = params[3]
	return
}

func (p *server) remove(key string) {
	category, service, err := p.path(key)
	if err != nil {
		log.Error(err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.datas[category]; ok {
		delete(p.datas[category], service)
	}
}

func (p *server) get(category, service string) (value int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.datas[category]; ok {
		value = p.datas[category][service]
	}

	return
}

func (p *server) get_group(category string) (group map[string]int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	group = p.datas[category]
	return
}

func (p *server) set(key, value string) {
	num, err := strconv.Atoi(value)
	if err != nil {
		log.Errorf("Set %v = %v strconv.Atoi err %v", key, value, err)
		return
	}

	category, service, err := p.path(key)
	if err != nil {
		log.Error(err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.datas[category]; !ok {
		p.datas[category] = make(map[string]int)
	}

	p.datas[category][service] = num
}

func (p *server) execute(f func(map[string]map[string]int)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	f(p.datas)
}

func (p *server) update(key, value string) {
	kAPI := etcdclient.NewKeysAPI(p.client)

	_, err := kAPI.Set(context.Background(), key, value, nil)
	if err != nil {
		log.Errorf("kapi set %v=%v err %v", key, value, err)
	}
}

func (p *server) load() {
	kAPI := etcdclient.NewKeysAPI(p.client)

	resp, err := kAPI.Get(context.Background(), p.root, &etcdclient.GetOptions{Recursive: true})
	if err != nil {
		log.Error(err)
		return
	}

	if !resp.Node.Dir {
		return
	}

	for _, node := range resp.Node.Nodes {
		if node.Dir {
			for _, v := range node.Nodes {
				p.set(v.Key, v.Value)
			}
		}
	}

	go p.watcher()
}

func (p *server) watcher() {
	kAPI := etcdclient.NewKeysAPI(p.client)
	w := kAPI.Watcher(p.root, &etcdclient.WatcherOptions{Recursive: true})
	for {
		resp, err := w.Next(context.Background())
		if err != nil {
			log.Error(err)
			continue
		}

		if resp.Node.Dir {
			continue
		}

		switch resp.Action {
		case "set", "create", "update", "compareAndSwap":
			p.set(resp.Node.Key, resp.Node.Value)
		case "delete":
			p.remove(resp.PrevNode.Key)
		}
	}
}

func ServiceState(category, service string) int {
	return _default_server.get(category, service)
}

func ServiceGroup(category string) map[string]int {
	return _default_server.get_group(category)
}

func SetServiceState(key, value string) {
	_default_server.update(_default_server.root+"/"+key+"/"+_default_server.service_id, value)
}
