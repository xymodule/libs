package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"gopkg.in/olivere/elastic.v6"
	"reflect"
)

var (
	client    *elastic.Client
	databases map[string]string
)

func Init(description map[string]string, host string) {
	if client != nil {
		return
	}
	databases = description
	var err error
	client, err = elastic.NewClient(elastic.SetURL(host))
	if err != nil {
		panic(err)
	}
	info, code, err := client.Ping(host).Do(context.Background())
	if err != nil {
		panic(err)
	}
	log.Debugf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, err := client.ElasticsearchVersion(host)
	if err != nil {
		panic(err)
	}
	log.Debugf("Elasticsearch version %s\n", esversion)
}

func NewBoolQuery(must map[string]string, should map[string]string) *elastic.BoolQuery {
	boolQ := elastic.NewBoolQuery()
	for k, v := range must {
		boolQ.Must(elastic.NewMatchQuery(k, v))
	}
	for k, v := range should {
		boolQ.Should(elastic.NewMatchQuery(k, v))
	}
	return boolQ
}

func Create(key string, id int64, message proto.Message) error {
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic create key %v not exist in databases", key)
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
		return nil, fmt.Errorf("elastic query key %v not exist in databases", key)
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
		return fmt.Errorf("elastic del key %v not exist in databases", key)
	}
	_, err := client.Delete().Index(key).
		Type(tpy).
		Id(fmt.Sprintf("%v", id)).
		Do(context.Background())
	return err
}

func Update(key string, id int64, data map[string]interface{}) error {
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic update key %v not exist in databases", key)
	}
	_, err := client.Update().
		Index(key).
		Type(tpy).
		Id(fmt.Sprintf("%v", id)).
		Doc(data).
		Do(context.Background())
	return err
}

func SortBy(key string, cond *elastic.BoolQuery, page int, size int, field string, asc bool) (items []interface{}, err error) {
	tpy, ok := databases[key]
	if !ok {
		return nil, fmt.Errorf("elastic sort key %v not exist in databases", key)
	}
	res, err := client.Search().
		Index(key).
		Type(tpy).
		Query(cond).
		Size(size).
		From((page - 1) * size).
		SortBy(elastic.NewFieldSort(field).Order(asc)).Do(context.Background())
	if err != nil {
		return nil, err
	}
	msgType := proto.MessageType(tpy)
	for _, item := range res.Each(msgType) {
		items = append(items, item)
	}
	return items, err
}

func BoolQuery(key string, cond *elastic.BoolQuery, size int) (items []interface{}, err error) {
	tpy, ok := databases[key]
	if !ok {
		return nil, fmt.Errorf("elastic bool query key %v not exist in databases", key)
	}
	res, err := client.Search(key).Type(tpy).Query(cond).Size(size).Do(context.Background())
	msgType := proto.MessageType(tpy)
	for _, item := range res.Each(msgType) {
		items = append(items, item)
	}
	return items, err
}

func Query(key string, page int, size int) (items []interface{}, err error) {
	tpy, ok := databases[key]
	if !ok {
		return nil, fmt.Errorf("elastic query list key %v not exist in databases", key)
	}
	var result *elastic.SearchResult
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

func Clear(key string) error {
	tpy, ok := databases[key]
	if !ok {
		return fmt.Errorf("elastic clear index %v not exist in databases", key)
	}
	_, err := client.Delete().Index(key).
		Type(tpy).
		Do(context.Background())
	return err
}
