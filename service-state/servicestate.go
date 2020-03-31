package servicestate

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	etcdclient "github.com/coreos/etcd/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	NUMBER_PREFIX_NODE = "number_prefixs"
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

func path_join(params ...string) string {
	return strings.Join(params, "/")
}

func path_dir(path string) string {
	dir := filepath.Dir(path)
	return strings.ReplaceAll(dir, "\\", "/")
}

type server struct {
	root           string
	service_id     string
	client         etcdclient.Client
	number_prefixs map[string]bool
	number_datas   map[string]map[string]int
	string_datas   map[string]map[string]string
	pathdatas      map[string]string
	callbacks      map[string][]chan string // callback on change
	mu             sync.RWMutex
}

func (p *server) init(root_path, service_id string, etcd_hosts []string) {
	p.root = root_path
	p.service_id = service_id

	p.number_prefixs = make(map[string]bool)
	p.number_datas = make(map[string]map[string]int)
	p.string_datas = make(map[string]map[string]string)
	p.pathdatas = make(map[string]string)
	p.callbacks = make(map[string][]chan string)

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

	p.load_number_prefixs(path_join(p.root, NUMBER_PREFIX_NODE))

	//
	p.load()
}

func (p *server) is_number_type(name string) bool {
	for k := range p.number_prefixs {
		if strings.HasPrefix(name, k) {
			return true
		}
	}

	return false
}

func (p *server) path(key string) (category, service string, err error) {
	var seperator string = "/"
	if strings.Contains(key, "\\") {
		seperator = "\\"
	}

	params := strings.Split(key, seperator)
	if len(params) != 4 {
		err = fmt.Errorf("Split %v len not equal 4", key)
		return
	}

	category = params[2]
	service = params[3]
	return
}

func (p *server) set(key, value string) {
	category, service, err := p.path(key)
	if err != nil {
		log.Error(err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.pathdatas[key] = value
	if ok := p.is_number_type(category); ok {
		num, err := strconv.Atoi(value)
		if err != nil {
			log.Errorf("Set %v = %v strconv.Atoi err %v", key, value, err)
			return
		}

		if _, ok := p.number_datas[category]; !ok {
			p.number_datas[category] = make(map[string]int)
		}

		p.number_datas[category][service] = num
	} else {
		if _, ok := p.string_datas[category]; !ok {
			p.string_datas[category] = make(map[string]string)
		}

		p.string_datas[category][service] = value
	}

	callback_path := path_dir(key)
	for k := range p.callbacks[callback_path] {
		select {
		case p.callbacks[callback_path][k] <- key:
		default:
		}
	}
}

func (p *server) remove(key string) {
	category, service, err := p.path(key)
	if err != nil {
		log.Error(err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if ok := p.is_number_type(category); ok {
		if _, ok := p.number_datas[category]; ok {
			delete(p.number_datas[category], service)
		}
	} else {
		if _, ok := p.string_datas[category]; ok {
			delete(p.number_datas[category], service)
		}
	}
	delete(p.pathdatas, key)

	callback_path := path_dir(key)
	for k := range p.callbacks[callback_path] {
		select {
		case p.callbacks[callback_path][k] <- key:
		default:
		}
	}
}

func (p *server) get_int(category, service string) (value int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ok := p.is_number_type(category); ok {
		if _, ok := p.number_datas[category]; ok {
			value = p.number_datas[category][service]
			return
		}
	}

	return -1
}

func (p *server) get_str(category, service string) (value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ok := p.is_number_type(category); !ok {
		if _, ok := p.string_datas[category]; ok {
			value = p.string_datas[category][service]
			return
		}
	}

	return
}

func (p *server) exec_group_int(category string, f func(map[string]int)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var group map[string]int
	if ok := p.is_number_type(category); ok {
		group = p.number_datas[category]
	}

	f(group)

	return
}

func (p *server) exec_group_str(category string, f func(map[string]string)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var group map[string]string
	if ok := p.is_number_type(category); !ok {
		group = p.string_datas[category]
	}

	f(group)

	return
}

func (p *server) update(key, value string) (err error) {
	kAPI := etcdclient.NewKeysAPI(p.client)

	var resp *etcdclient.Response
	resp, err = kAPI.Get(context.Background(), key, nil)
	if err != nil {
		_, err = kAPI.Create(context.Background(), key, value)
		return
	}

	if value == resp.Node.Value {
		return fmt.Errorf("update %v, %v duplicated", key, value)
	}

	prevIndex := resp.Node.ModifiedIndex
	_, err = kAPI.Set(context.Background(), key, value, &etcdclient.SetOptions{PrevIndex: prevIndex})
	return
}

func (p *server) load_number_prefixs(filepath string) {
	kAPI := etcdclient.NewKeysAPI(p.client)

	// get the keys under directory
	log.Infof("reading number types from:%v", filepath)
	resp, err := kAPI.Get(context.Background(), filepath, nil)
	if err != nil {
		log.Error(err)
		return
	}

	// validation check
	if resp.Node.Dir {
		log.Error("types is not a node")
		return
	}

	// split types
	types := strings.Split(resp.Node.Value, " ")
	for _, v := range types {
		p.number_prefixs[v] = true
	}

	log.Infof("reading number types :%v", resp.Node.Value)
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

func (p *server) register_callback(path string, callback chan string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.callbacks[path] = append(p.callbacks[path], callback)

	for k := range p.pathdatas {
		callback_path := path_dir(k)
		if callback_path != path {
			continue
		}

		callback <- k
	}
	log.Infof("register callback on: %v", path)
}

func ServiceVarInt(category, service string) int {
	return _default_server.get_int(category, service)
}

func ServiceVarStr(category, service string) string {
	return _default_server.get_str(category, service)
}

func ExecuteGroupInt(category string, f func(map[string]int)) {
	_default_server.exec_group_int(category, f)
}

func ExecuteGroupStr(category string, f func(map[string]string)) {
	_default_server.exec_group_str(category, f)
}

func SetServiceVar(path, value string) {
	UpdateGlobalVar(path, _default_server.service_id, value)
}

func UpdateGlobalVar(path, key, value string) {
	_default_server.update(path_join(_default_server.root, path, key), value)
}

func RegisterCallback(path string, callback chan string) {
	_default_server.register_callback(path_join(_default_server.root, path), callback)
}
