package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/storage/redis/v3"
)

type Config redis.Config

type Store struct {
	client *redis.Storage
}

// New 생성자
func New(config ...Config) *Store {
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
		c := config[0]
		if c.Host != "" {
			options.Host = c.Host
		}
		if c.Port != 0 {
			options.Port = c.Port
		}
		if c.Database != 0 {
			options.Database = c.Database
		}
		if c.Password != "" {
			options.Password = c.Password
		}
		if c.Username != "" {
			options.Username = c.Username
		}
	}

	return &Store{client: redis.New(options)}
}

// Set 값 저장 (JSON 직렬화)
func (s *Store) Set(key string, value any, ttl ...time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	exp := time.Duration(0)
	if len(ttl) > 0 {
		exp = ttl[0]
	}
	return s.client.Set(key, jsonData, exp)
}

// Get 값 조회 (구조체 언마샬링)
func (s *Store) Get(key string, dest any) error {
	value, err := s.client.Get(key)
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if len(value) == 0 {
		return errors.New("key not found")
	}
	return json.Unmarshal(value, dest)
}

func (s *Store) GetString(key string) (string, error) {
	var data string
	err := s.Get(key, &data)
	return data, err
}

func (s *Store) GetInt(key string) (int, error) {
	var data int
	err := s.Get(key, &data)
	return data, err
}

func (s *Store) GetFloat(key string) (float64, error) {
	var data float64
	err := s.Get(key, &data)
	return data, err
}

func (s *Store) GetBool(key string) (bool, error) {
	var data bool
	err := s.Get(key, &data)
	return data, err
}

// Delete 키 삭제
func (s *Store) Delete(key string) error {
	return s.client.Delete(key)
}

// Reset 전체 삭제 (안전 장치 추가)
func (s *Store) Reset(confirm bool) error {
	if !confirm {
		return errors.New("reset not confirmed")
	}
	return s.client.Reset()
}

// Close 종료
func (s *Store) Close() {
	s.client.Close()
}
