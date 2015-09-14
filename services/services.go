package services

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/go-etcd/etcd"
	log "github.com/gonet2/libs/nsq-logger"
	"google.golang.org/grpc"
)

const (
	DEFAULT_ETCD         = "http://127.0.0.1:2379"
	DEFAULT_SERVICE_PATH = "/backends"
	SERVICE_ACCESS       = "/service_access"
	DEFAULT_DIAL_TIMEOUT = 10 * time.Second
	RETRY_DELAY          = 10 * time.Second
)

type client struct {
	key  string
	conn *grpc.ClientConn
}

type service struct {
	clients []client
	idx     uint32
}

type service_pool struct {
	services    map[string]*service
	client_pool sync.Pool
	sync.RWMutex

	name   string
	access map[string]bool
}

var (
	_default_pool service_pool
)

func init() {
	_default_pool.init()
	_default_pool.init_access()
	_default_pool.connect_all(DEFAULT_SERVICE_PATH)
	go _default_pool.watcher()
}

func (p *service_pool) init() {
	if env := os.Getenv("SERVICE_NAME"); env == "" {
		log.Error("SERVICE_NAME is not set.")
		return
	} else {
		p.name = env
	}

	// etcd client
	machines := []string{DEFAULT_ETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}
	p.client_pool.New = func() interface{} {
		return etcd.NewClient(machines)
	}

	p.services = make(map[string]*service)
}

// access must be stored like /service_access/xxx_service yyy_service:1,zzz_service:1
func (p *service_pool) init_access() {
	client := p.client_pool.Get().(*etcd.Client)
	defer func() {
		p.client_pool.Put(client)
	}()

	resp, err := client.Get(SERVICE_ACCESS+"/"+p.name, false, false)
	if err != nil {
		log.Error(err)
		return
	}
	p.access = make(map[string]bool)
	for _, service := range resp.Node.Nodes {
		if service.Value == "1" {
			_, name := filepath.Split(service.Key)
			p.access[DEFAULT_SERVICE_PATH+"/"+name] = true
		}
	}
}

// connect to all services
func (p *service_pool) connect_all(directory string) {
	client := p.client_pool.Get().(*etcd.Client)
	defer func() {
		p.client_pool.Put(client)
	}()

	// get the keys under directory
	log.Info("connecting services under:", directory)
	resp, err := client.Get(directory, true, true)
	if err != nil {
		log.Error(err)
		return
	}

	// validation check
	if !resp.Node.Dir {
		log.Error("not a directory")
		return
	}
	for _, node := range resp.Node.Nodes {
		if node.Dir { // service directory
			for _, service := range node.Nodes {
				p.add_service(service.Key, service.Value)
			}
		} else {
			log.Warning("malformed service directory:", node.Key)
		}
	}
	log.Info("services add complete")
}

// watcher for data change in etcd directory
func (p *service_pool) watcher() {
	client := p.client_pool.Get().(*etcd.Client)
	defer func() {
		p.client_pool.Put(client)
	}()

	for {
		ch := make(chan *etcd.Response, 10)
		go func() {
			for {
				if resp, ok := <-ch; ok {
					if resp.Node.Dir {
						continue
					}
					key, value := resp.Node.Key, resp.Node.Value
					if value == "" {
						p.remove_service(key)
					} else {
						p.add_service(key, value)
					}
				} else {
					return
				}
			}
		}()

		_, err := client.Watch(DEFAULT_SERVICE_PATH, 0, true, ch, nil)
		if err != nil {
			log.Critical(err)
		}
		<-time.After(RETRY_DELAY)
	}
}

// add a service
func (p *service_pool) add_service(key, value string) {
	p.Lock()
	defer p.Unlock()
	service_name := filepath.Dir(key)
	if _, ok := p.access[service_name]; !ok {
		return
	}
	if p.services[service_name] == nil {
		p.services[service_name] = &service{}
		log.Tracef("new service type: %v", service_name)
	}
	service := p.services[service_name]

	if conn, err := grpc.Dial(value, grpc.WithTimeout(DEFAULT_DIAL_TIMEOUT)); err == nil {
		service.clients = append(service.clients, client{key, conn})
		log.Tracef("service added: %v -- %v", key, value)
	} else {
		log.Errorf("did not connect: %v -- %v err: %v", key, value, err)
	}
}

// remove a service
func (p *service_pool) remove_service(key string) {
	p.Lock()
	defer p.Unlock()
	service_name := filepath.Dir(key)
	if _, ok := p.access[service_name]; !ok {
		return
	}
	service := p.services[service_name]
	if service == nil {
		log.Tracef("no such service %v", service_name)
		return
	}

	for k := range service.clients {
		if service.clients[k].key == key { // deletion
			service.clients[k].conn.Close()
			service.clients = append(service.clients[:k], service.clients[k+1:]...)
			log.Tracef("service removed %v", key)
			return
		}
	}
}

// provide a specific key for a service, eg:
// path:/backends/snowflake, id:s1
//
// service must be stored like /backends/xxx_service/xxx_id
func (p *service_pool) get_service_with_id(path string, id string) *grpc.ClientConn {
	p.RLock()
	defer p.RUnlock()
	service := p.services[path]
	if service == nil || len(service.clients) == 0 {
		return nil
	}

	fullpath := path + "/" + id
	for k := range service.clients {
		if service.clients[k].key == fullpath {
			return service.clients[k].conn
		}
	}

	return nil
}

func (p *service_pool) get_service(path string) *grpc.ClientConn {
	p.RLock()
	defer p.RUnlock()
	service := p.services[path]
	if service == nil {
		return nil
	}

	if len(service.clients) == 0 {
		return nil
	}
	idx := int(atomic.AddUint32(&service.idx, 1))
	return service.clients[idx%len(service.clients)].conn
}

// choose a service randomly
func GetService(path string) *grpc.ClientConn {
	return _default_pool.get_service(path)
}

// get a specific service instance with given path and id
func GetServiceWithId(path string, id string) *grpc.ClientConn {
	return _default_pool.get_service_with_id(path, id)
}
