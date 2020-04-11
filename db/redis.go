package db

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/redis.v3"
)

var (
	_default_redis *redis.ClusterClient
)

func InitRedis(hosts []string, pass string, poolsize int) {
	log.Infof("redis hosts %v, pass %v", hosts, pass)
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    hosts,
		Password: pass,
		PoolSize: poolsize,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Errorf("Connect to redis servers failed: %v", err)
		os.Exit(-1)
	} else {
		log.Info(pong)
	}

	_default_redis = client
}

func GetRedis() *redis.ClusterClient {
	return _default_redis
}

func IsNilReply(err error) bool {
	return err == redis.Nil
}

func RedisSet(key string, data interface{}) (err error) {
	client := GetRedis()
	if client == nil {
		err = errors.New("redis service unavailable")
		return
	}

	reply := client.Set(key, data, 0)
	err = reply.Err()
	return
}

func RedisGet(key string) (reply *redis.StringCmd, err error) {
	client := GetRedis()
	if client == nil {
		err = errors.New("redis service unavailable")
		return
	}

	reply = client.Get(key)
	err = reply.Err()
	return
}

func RedisHSet(hKey, key, data string) (success bool, err error) {
	client := GetRedis()
	if client == nil {
		err = errors.New("redis service unavailable")
		return
	}

	success, err = client.HSet(hKey, key, data).Result()
	return
}

func RedisHGet(hKey, key string) (value string, err error) {
	client := GetRedis()
	if client == nil {
		err = errors.New("redis service unavailable")
		return
	}

	value, err = client.HGet(hKey, key).Result()
	return
}

func RedisHMGet(hKey string, keys ...string) (values []interface{}, err error) {
	client := GetRedis()
	if client == nil {
		err = errors.New("redis service unavailable")
		return
	}

	values, err = client.HMGet(hKey, keys...).Result()
	return
}

func RedisHIncr(hKey, key string, num int64) (err error) {
	client := GetRedis()
	if client == nil {
		err = errors.New("redis service unavailable")
		return
	}

	reply := client.HIncrBy(hKey, key, num)
	err = reply.Err()
	return
}

func RedisHash(hKey string) (hash map[string]string, err error) {
	client := GetRedis()
	if client == nil {
		err = errors.New("redis service unavailable")
		return
	}

	hash, err = client.HGetAllMap(hKey).Result()
	return
}
