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
)

func Init(description map[string]string, nodes ...string) {
	if client != nil {
		return
	}
	databases = description
	var err error
	var options []els.ClientOptionFunc
	options = append(options, els.SetURL(nodes...))
	options = append(options, els.SetSniff(false))
	options = append(options, els.SetHealthcheck(true))
	options = append(options, els.SetHealthcheckInterval(10*time.Second))
	options = append(options, els.SetErrorLog(log.New()))
	client, err = els.NewClient(options...)
	if err != nil {
		panic(err)
	}
	info, code, err := client.Ping(nodes[0]).Do(context.Background())
	if err != nil {
		panic(err)
	}
	log.Infof("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, err := client.ElasticsearchVersion(nodes[0])
	if err != nil {
		panic(err)
	}
	log.Infof("Elasticsearch version %s\n", esversion)
	//PutMapping().Index(testIndexName3).Type("doc").BodyString(mapping).Do(context.TODO())
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
	_, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic del by query key %v not exist in indices", key)
	}
	_, err := client.DeleteByQuery(key).Query(cond).Do(context.Background())
	return err
}

func SaveOrUpdate(key string, id int64, message proto.Message) error {
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
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic clear index %v not exist in indices", key)
	}
	_, err := client.Delete().Index(key).
		Type(tpy).
		Do(context.Background())
	return err
}
