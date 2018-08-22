package db

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

type ProtoByte struct {
	base interface{}
	key  string
}

func (p *ProtoByte) Bind(o interface{}, key string) {
	p.base = o
	p.key = key
}

func (p *ProtoByte) Load() error {
	reply, err := RedisGet(p.key)
	if err != nil {
		return fmt.Errorf("Load data %v redisGet error %v", p.key, err)
	}

	bytes, err := reply.Bytes()
	if err != nil {
		return fmt.Errorf("Load data %v reply error %v", p.key, err)
	}

	err = proto.Unmarshal(bytes, p.base.(proto.Message))
	if err != nil {
		return fmt.Errorf("Load data %v unmarshal error %v", p.key, err)
	}

	return nil
}

func (p *ProtoByte) Save() error {
	bytes, err := proto.Marshal(p.base.(proto.Message))
	if err != nil {
		return fmt.Errorf("Save data %v marshal error %v", p.key, err)
	}

	err = RedisSet(p.key, bytes)
	if err != nil {
		return fmt.Errorf("Save data %v redisSet error %v", p.key, err)
	}

	return nil
}

type ProtoHash struct {
	base interface{}
	hKey string
}

func (p *ProtoHash) Bind(o interface{}, key string) {
	p.base = o
	p.hKey = key
}

func (p *ProtoHash) Load() (err error) {
	ov := reflect.ValueOf(p.base)
	if ov.IsNil() {
		err = errors.New("Load base value is nil")
		return
	}

	hashmap, err := RedisHash(p.hKey)
	if err != nil {
		return err
	}

	for key, value := range hashmap {
		fv := ov.Elem().FieldByName(key)
		if !fv.IsValid() {
			log.Errorf("Load field %v is invalid", key)
			// 兼容字段修改
			continue
		}

		err = p.parse(fv, value)
		if err != nil {
			err = fmt.Errorf("parse field %v err %v", key, err)
			return
		}
	}

	return
}

func (p *ProtoHash) Save() error {
	rf := reflect.ValueOf(p.base)
	ov := reflect.Indirect(rf)

	for i := 0; i < ov.NumField(); i++ {
		name := ov.Type().Field(i).Name
		if name == "XXX_unrecognized" {
			continue
		}

		err := p.SaveField(name)
		if err != nil {
			log.Error(err)
			continue
		}
	}

	return nil
}

func (p *ProtoHash) LoadField(name string) (err error) {
	ov := reflect.ValueOf(p.base)
	if ov.IsNil() {
		err = fmt.Errorf("LoadField %v base value is nil", name)
		return
	}

	fv := ov.Elem().FieldByName(name)
	if !fv.IsValid() {
		err = fmt.Errorf("LoadField %v field is invalid", name)
		return
	}

	var value string
	value, err = RedisHGet(p.hKey, name)
	if err != nil {
		if IsNilReply(err) {
			err = nil
			return
		}

		err = fmt.Errorf("LoadField %v error %v", name, err)
		return
	}

	err = p.parse(fv, value)
	if err != nil {
		log.Errorf("parse field %v error %v", name, err)
		return err
	}

	return nil
}

func (p *ProtoHash) SaveField(name string) (err error) {
	ov := reflect.ValueOf(p.base)
	if ov.IsNil() {
		err = fmt.Errorf("SaveField %v base value is nil", name)
		return
	}

	fv := ov.Elem().FieldByName(name)
	if !fv.IsValid() {
		err = fmt.Errorf("SaveField %v field is invalid", name)
		return
	}

	if fv.IsNil() {
		err = fmt.Errorf("SaveField field %v value is nil", name)
		return
	}

	var value string
	value, err = p.format(fv)
	if err != nil {
		return
	}

	_, err = RedisHSet(p.hKey, name, value)
	if err != nil {
		return
	}

	return nil
}

func (p *ProtoHash) parse(fieldValue reflect.Value, value string) error {
	ft := fieldValue.Type()
	if strings.Contains(ft.String(), "proto.") {
		if fieldValue.IsNil() {
			// New one
			v := reflect.New(ft.Elem())
			fieldValue.Set(v)
		}

		err := proto.Unmarshal([]byte(value), fieldValue.Interface().(proto.Message))
		if err != nil {
			return err
		}

		return nil
	}

	switch ft.String() {
	case "*int32":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(proto.Int32(int32(v))))
	case "*int64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(proto.Int64(v)))
	case "*uint32":
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(proto.Uint32(uint32(v))))
	case "*uint64":
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(proto.Uint64(v)))
	case "*float32":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(proto.Float32(float32(v))))
	case "*float64":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(proto.Float64(float64(v))))
	case "*bytes":
		fallthrough
	case "*string":
		fieldValue.Set(reflect.ValueOf(proto.String(value)))
	case "*bool":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(proto.Bool(v)))
	case "enum":
		//TODO
		/*
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			en := int32(v)
			//fieldValue.Set(reflect.ValueOf(proto.Int32(int32(v))))
		*/
	default:
		return errors.New("the type of value not recogized")
	}

	return nil
}

func (p *ProtoHash) format(fieldValue reflect.Value) (value string, err error) {
	ft := fieldValue.Type()
	if strings.Contains(ft.String(), "proto.") {
		if !fieldValue.IsNil() {
			var bytes []byte
			bytes, err = proto.Marshal(fieldValue.Interface().(proto.Message))
			if err != nil {
				return
			}

			value = string(bytes)
		}

		return
	}

	switch ft.String() {
	case "*int32":
		v := fieldValue.Interface().(*int32)
		value = strconv.FormatInt(int64(*v), 10)
		return
	case "*int64":
		v := fieldValue.Interface().(*int64)
		value = strconv.FormatInt(*v, 10)
		return
	case "*uint32":
		v := fieldValue.Interface().(*uint32)
		value = strconv.FormatUint(uint64(*v), 10)
		return
	case "*uint64":
		v := fieldValue.Interface().(*uint64)
		value = strconv.FormatUint(*v, 10)
		return
	case "*float32":
		v := fieldValue.Interface().(*float32)
		value = strconv.FormatFloat(float64(*v), 'E', -1, 32)
		return
	case "*float64":
		v := fieldValue.Interface().(*float64)
		value = strconv.FormatFloat(*v, 'E', -1, 64)
		return
	case "*bytes":
		v := fieldValue.Interface().(*string)
		value = *v
		return
	case "*string":
		v := fieldValue.Interface().(*string)
		value = *v
		return
	case "*bool":
		v := fieldValue.Interface().(*bool)
		value = strconv.FormatBool(*v)
		return
	default:
		err = errors.New("type not recognized")
	}

	return
}
