package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	els "github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

var (
	client    *els.Client
	databases map[string]string
	hosts     []string
)

func Init(description map[string]string, nodes ...string) {
	if client != nil {
		return
	}
	databases = description
	hosts = nodes
	var err error
	var options []els.ClientOptionFunc
	options = append(options, els.SetURL(nodes...))
	options = append(options, els.SetSniff(false))
	options = append(options, els.SetHealthcheck(true))
	options = append(options, els.SetHealthcheckInterval(10*time.Second))
	options = append(options, els.SetRetrier(els.NewBackoffRetrier(els.NewSimpleBackoff(100, 100, 100, 100, 100))))
	options = append(options, els.SetErrorLog(log.New()))
	client, err = els.NewClient(options...)
	if err != nil {
		log.Panicf("Elasticsearch err %v", err)
	}
	info, code, err := client.Ping(nodes[0]).Do(context.Background())
	if err != nil {
		log.Panicf("Elasticsearch err %v", err)
	}
	log.Infof("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, err := client.ElasticsearchVersion(nodes[0])
	if err != nil {
		log.Panicf("Elasticsearch err %v", err)
	}
	log.Infof("Elasticsearch version %s\n", esversion)
	go func() {
		defer recovery()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := checkHealth(); err != nil {
					log.Errorf("Elasticsearch checkHealth err %v", err)
					err = reconnect()
					if err != nil {
						log.Errorf("Elasticsearch reconnect err %v", err)
					}
				}
			}
		}
	}()
	//PutMapping().Index(testIndexName3).Type("doc").BodyString(mapping).Do(context.TODO())
}

func recovery() {
	if r := recover(); r != nil {
		log.Errorf("Elasticsearch recovered:", r)
	}
}

func reconnect() error {
	var err error
	if client != nil {
		client.Stop()
	}

	var options []els.ClientOptionFunc
	options = append(options, els.SetURL(hosts...))
	options = append(options, els.SetSniff(false))
	options = append(options, els.SetHealthcheck(true))
	options = append(options, els.SetHealthcheckInterval(10*time.Second))
	options = append(options, els.SetRetrier(els.NewBackoffRetrier(els.NewSimpleBackoff(100, 100, 100, 100, 100))))
	options = append(options, els.SetErrorLog(log.New()))
	client, err = els.NewClient(options...)
	if err != nil {
		return err
	}
	return nil
}

func checkHealth() error {
	if client == nil {
		return fmt.Errorf("elastic checkHealth client is nil")
	}
	_, err := client.CatHealth().Do(context.Background())
	return err
}

func CreateMapping(key string, mapping string) error {
	_, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic new mapping key %v not exist in indices", key)
	}
	exist, err := client.IndexExists(key).Do(context.Background())
	if err != nil {
		return err
	}
	if !exist {
		_, err = client.CreateIndex(key).BodyString(mapping).Do(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

func NewBoolQuery(must map[string]string, should map[string]string) *els.BoolQuery {
	boolQ := els.NewBoolQuery()
	for k, v := range must {
		boolQ.Must(els.NewMatchQuery(k, v))
	}
	for k, v := range should {
		boolQ.Should(els.NewMatchQuery(k, v))
	}
	return boolQ
}

func Create(key string, id int64, message proto.Message) error {
	if client == nil {
		return fmt.Errorf("elastic Create client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic create key %v not exist in indices", key)
	}
	_, err := client.Index().
		Index(key).
		Type(tpy).
		Id(fmt.Sprintf("%v", id)).
		BodyJson(message).
		Do(context.Background())
	return err
}

func Get(key string, id int64) (interface{}, error) {
	if client == nil {
		return nil, fmt.Errorf("elastic Get client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return nil, fmt.Errorf("elastic query key %v not exist in indices", key)
	}
	doc, err := client.Get().Index(key).Type(tpy).Id(fmt.Sprintf("%v", id)).Do(context.Background())
	if err != nil {
		return nil, err
	}
	bytes, err := doc.Source.MarshalJSON()
	if err != nil {
		return nil, err
	}
	message := newMessageByName(tpy)
	err = json.Unmarshal(bytes, message)
	return message, err
}

func newMessageByName(name string) interface{} {
	mType := proto.MessageType(name)
	return reflect.New(mType.Elem()).Interface()
}

func Delete(key string, id int64) error {
	if client == nil {
		return fmt.Errorf("elastic Delete client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic del key %v not exist in indices", key)
	}
	_, err := client.Delete().Index(key).
		Type(tpy).
		Id(fmt.Sprintf("%v", id)).
		Do(context.Background())
	return err
}

func DeleteByQuery(key string, cond *els.BoolQuery) error {
	if client == nil {
		return fmt.Errorf("elastic DeleteByQuery client is nil")
	}
	_, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic del by query key %v not exist in indices", key)
	}
	_, err := client.DeleteByQuery(key).Query(cond).Do(context.Background())
	return err
}

func SaveOrUpdate(key string, id int64, message proto.Message) error {
	if client == nil {
		return fmt.Errorf("elastic SaveOrUpdate client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic update key %v not exist in indices", key)
	}
	_, err := client.Update().
		Index(key).
		Type(tpy).
		Id(fmt.Sprintf("%v", id)).
		Doc(message).
		DocAsUpsert(true).
		Do(context.Background())
	return err
}

func Update(key string, id int64, message proto.Message) error {
	if client == nil {
		return fmt.Errorf("elastic Update client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic update key %v not exist in indices", key)
	}
	_, err := client.Update().
		Index(key).
		Type(tpy).
		Id(fmt.Sprintf("%v", id)).
		Doc(message).
		Do(context.Background())
	return err
}

func SortBy(key string, cond *els.BoolQuery, page int, size int, field string, asc bool) (items []interface{}, err error) {
	if client == nil {
		return nil, fmt.Errorf("elastic SortBy client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return nil, fmt.Errorf("elastic sort key %v not exist in indices", key)
	}
	res, err := client.Search().
		Index(key).
		Type(tpy).
		Query(cond).
		Size(size).
		From((page - 1) * size).
		SortBy(els.NewFieldSort(field).Order(asc)).Do(context.Background())
	if err != nil {
		return nil, err
	}
	msgType := proto.MessageType(tpy)
	for _, item := range res.Each(msgType) {
		items = append(items, item)
	}
	return items, err
}

func BoolQuery(key string, cond *els.BoolQuery, size int) (items []interface{}, err error) {
	if client == nil {
		return nil, fmt.Errorf("elastic BoolQuery client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return nil, fmt.Errorf("elastic bool query key %v not exist in indices", key)
	}
	res, err := client.Search(key).Type(tpy).Query(cond).Size(size).Do(context.Background())
	if err != nil {
		return nil, err
	}
	msgType := proto.MessageType(tpy)
	for _, item := range res.Each(msgType) {
		items = append(items, item)
	}
	return items, err
}

func Query(key string, page int, size int) (items []interface{}, err error) {
	if client == nil {
		return nil, fmt.Errorf("elastic Query client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return nil, fmt.Errorf("elastic query list key %v not exist in indices", key)
	}
	var result *els.SearchResult
	if size > 0 {
		result, err = client.Search(key).Type(tpy).Size(size).From((page - 1) * size).Do(context.Background())
	} else {
		result, err = client.Search(key).Type(tpy).Do(context.Background())
	}
	if err != nil {
		return nil, err
	}
	msgType := proto.MessageType(tpy)
	for _, item := range result.Each(msgType) {
		items = append(items, item)
	}
	return items, nil
}

func Exist(key string) bool {
	if client == nil {
		log.Errorf("elastic Exist client is nil")
		return false
	}
	_, ok := databases[key]
	if !ok {
		return false
	}
	exists, err := client.IndexExists(key).Do(context.Background())
	if err != nil {
		return false
	}
	return exists
}

func Clear(key string) error {
	if client == nil {
		return fmt.Errorf("elastic Clear client is nil")
	}
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic clear index %v not exist in indices", key)
	}
	_, err := client.Delete().Index(key).
		Type(tpy).
		Do(context.Background())
	return err
}
