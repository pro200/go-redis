package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/storage/redis/v3"
)

type Config struct {
	Name     string // database 이름
	Host     string
	Port     int
	Database int
	Username string
	Password string
}

type Database struct {
	client *redis.Storage
}

var Databases = make(map[string]*Database)

// New 생성자
func New(config ...Config) *Database {
	dbName := "main"
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
		if c.Name != "" {
			dbName = c.Name
		}
		if c.Host != "" {
			options.Host = c.Host
		}
		if c.Port != 0 {
			options.Port = c.Port
		}
		if c.Database != 0 {
			options.Database = c.Database
		}
		if c.Username != "" {
			options.Username = c.Username
		}
		if c.Password != "" {
			options.Password = c.Password
		}
	}

	Databases[dbName] = &Database{client: redis.New(options)}
	return Databases[dbName]
}

func GetDatabase(name ...string) (*Database, error) {
	dbName := "main"
	if len(name) > 0 {
		dbName = name[0]
	}
	db, ok := Databases[dbName]
	if !ok {
		return nil, fmt.Errorf("database %s not found", dbName)
	}
	return db, nil
}

// Set 값 저장 (JSON 직렬화)
func (d *Database) Set(key string, value any, ttl ...time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	exp := time.Duration(0)
	if len(ttl) > 0 {
		exp = ttl[0]
	}
	return d.client.Set(key, jsonData, exp)
}

// Get 값 조회 (구조체 언마샬링)
func (d *Database) Get(key string, dest any) error {
	value, err := d.client.Get(key)
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if len(value) == 0 {
		return errors.New("key not found")
	}
	return json.Unmarshal(value, dest)
}

func (d *Database) GetString(key string) (string, error) {
	var data string
	err := d.Get(key, &data)
	return data, err
}

func (d *Database) GetInt(key string) (int, error) {
	var data int
	err := d.Get(key, &data)
	return data, err
}

func (d *Database) GetFloat(key string) (float64, error) {
	var data float64
	err := d.Get(key, &data)
	return data, err
}

func (d *Database) GetBool(key string) (bool, error) {
	var data bool
	err := d.Get(key, &data)
	return data, err
}

// Delete 키 삭제
func (d *Database) Delete(key string) error {
	return d.client.Delete(key)
}

// Reset 전체 삭제 (안전 장치 추가)
func (d *Database) Reset(confirm bool) error {
	if !confirm {
		return errors.New("reset not confirmed")
	}
	return d.client.Reset()
}

// Close 종료
func (d *Database) Close() {
	d.client.Close()
}

func Close() {
	for _, database := range Databases {
		database.Close()
	}
}
