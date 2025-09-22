package redis

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/storage/redis/v3"
)

var storage *redis.Storage

type Config redis.Config

func Init(config ...Config) {
	options := redis.Config{
		Host:      "127.0.0.1",
		Port:      6379,
		Database:  0,
		Username:  "",
		Password:  "",
		Reset:     false,
		TLSConfig: nil,
		PoolSize:  10 * runtime.GOMAXPROCS(0),
	}

	if len(config) > 0 {
		if config[0].Host != "" {
			options.Host = config[0].Host
		}
		if config[0].Port != 0 {
			options.Port = config[0].Port
		}
		if config[0].Database != 0 {
			options.Database = config[0].Database
		}
		if config[0].Password != "" {
			options.Password = config[0].Password
		}
		if config[0].Username != "" {
			options.Username = config[0].Username
		}
	}

	storage = redis.New(options)
}

// Redis에 값을 JSON으로 변환하여 저장
func Set(key string, value interface{}, ttl ...time.Duration) error {
	// value를 JSON으로 직렬화
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// ttl이 없으면 기본값으로 0을 설정
	if len(ttl) == 0 {
		ttl = append(ttl, 0)
	}

	// Redis에 JSON 데이터를 저장
	return storage.Set(key, jsonData, ttl[0])
}

// Redis에서 JSON값을 가져와 dest구조체로 언마샬링
func Get(key string, dest interface{}) error {
	// Redis에서 데이터 가져오기
	value, err := storage.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get value from Redis: %w", err)
	}
	if len(value) == 0 {
		return fmt.Errorf("key not found")
	}

	// JSON 데이터를 구조체로 언마샬링
	if err = json.Unmarshal(value, dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func GetString(key string) (string, error) {
	var data string
	err := Get(key, &data)
	return data, err
}

func GetInt(key string) (int, error) {
	var data int
	err := Get(key, &data)
	return data, err
}

func GetFloat(key string) (float64, error) {
	var data float64
	err := Get(key, &data)
	return data, err
}

func GetBool(key string) (bool, error) {
	var data bool
	err := Get(key, &data)
	return data, err
}

func Reset() error {
	return storage.Reset()
}

func Delete(key string) error {
	return storage.Delete(key)
}

func Close() {
	_ = storage.Close()
}
