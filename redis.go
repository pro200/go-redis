package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
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
	client *redis.Client
}

var (
	Databases = make(map[string]*Database)
	ctx       = context.Background()
	dbMu      sync.RWMutex // 동시성 안전 보장
)

// New 생성자
func New(config ...Config) *Database {
	dbName := "main"
	options := &redis.Options{
		Addr:      "127.0.0.1:6379",
		DB:        0,
		Username:  "",
		Password:  "",
		TLSConfig: nil,
		PoolSize:  10 * runtime.GOMAXPROCS(0),
	}

	if len(config) > 0 {
		c := config[0]
		if c.Name != "" {
			dbName = c.Name
		}
		if c.Host != "" {
			options.Addr = c.Host + ":6379"
		}
		if c.Port != 0 {
			options.Addr = c.Host + fmt.Sprintf(":%d", c.Port)
		}
		if c.Database != 0 {
			options.DB = c.Database
		}
		if c.Username != "" {
			options.Username = c.Username
		}
		if c.Password != "" {
			options.Password = c.Password
		}
	}

	database := &Database{client: redis.NewClient(options)}

	dbMu.Lock()
	Databases[dbName] = database
	dbMu.Unlock()

	return database
}

func GetDatabase(name ...string) (*Database, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if len(Databases) == 0 {
		return nil, errors.New("no databases available")
	}

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

	return d.client.Set(ctx, key, jsonData, exp).Err()
}

// Get 값 조회 (구조체 언마샬링)
func (d *Database) Get(key string, dest any) error {
	value, err := d.client.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if len(value) == 0 {
		return errors.New("key not found")
	}
	return json.Unmarshal([]byte(value), dest)
}

// Push 리스트에 값 추가 (Left / Right)
func (d *Database) LPush(key string, value any) error {
	return d.push("Left", key, value)
}

func (d *Database) RPush(key string, value any) error {
	return d.push("Right", key, value)
}

func (d *Database) push(direction string, key string, value any) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if direction == "Left" || direction == "L" || direction == "l" {
		return d.client.LPush(ctx, key, jsonData).Err()
	}
	return d.client.RPush(ctx, key, jsonData).Err()
}

// Pop 리스트에서 값 추출 (Left / Right)
func (d *Database) LPop(key string, dest any) error {
	return d.pop("Left", key, dest)
}

func (d *Database) RPop(key string, dest any) error {
	return d.pop("Right", key, dest)
}

func (d *Database) pop(direction string, key string, dest any) error {
	var value string
	var err error

	if direction == "Left" || direction == "L" || direction == "l" {
		value, err = d.client.LPop(ctx, key).Result()
	} else {
		value, err = d.client.RPop(ctx, key).Result()
	}
	if err != nil {
		return fmt.Errorf("no items in the list \"%s\"", key)
	}

	return json.Unmarshal([]byte(value), dest)
}

// LLen 리스트 길이 조회
func (d *Database) LLen(key string) (int64, error) {
	return d.client.LLen(ctx, key).Result()
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
	return d.client.Del(ctx, key).Err()
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
