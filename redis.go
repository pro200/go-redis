package redis

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	//"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type Config struct {
	//Name     string
	Host     string
	Port     int
	Database int
	Username string
	Password string
}

type Database struct {
	client *redis.Client
	ctx    context.Context
}

//var (
//	Databases = make(map[string]*Database)
//	dbMu      sync.RWMutex
//)

// Host/Port 안전 포맷
//func NewRedisAddr(host string, port int) string {
//	if host == "" {
//		host = "127.0.0.1"
//	}
//	if port == 0 {
//		port = 6379
//	}
//	return fmt.Sprintf("%s:%d", host, port)
//}

// New: 생성자
func New(cfg ...Config) *Database {
	c := Config{
		//Name:     "main",
		Host:     "127.0.0.1",
		Port:     6379,
		Database: 0,
	}

	if len(cfg) > 0 {
		tmp := cfg[0]
		//if tmp.Name != "" {
		//	c.Name = tmp.Name
		//}
		if tmp.Host != "" {
			c.Host = tmp.Host
		}
		if tmp.Port != 0 {
			c.Port = tmp.Port
		}
		if tmp.Database != 0 {
			c.Database = tmp.Database
		}
		if tmp.Username != "" {
			c.Username = tmp.Username
		}
		if tmp.Password != "" {
			c.Password = tmp.Password
		}
	}

	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		DB:       c.Database,
		Username: c.Username,
		Password: c.Password,
		PoolSize: 10 * runtime.GOMAXPROCS(0),
	}

	//db := &Database{
	//	client: redis.NewClient(options),
	//	ctx:    context.Background(),
	//}
	//
	//dbMu.Lock()
	//Databases[c.Name] = db
	//dbMu.Unlock()
	//
	//return db

	return &Database{
		client: redis.NewClient(options),
		ctx:    context.Background(),
	}
}

// Load: 이름으로 DB 가져오기
//func Load(name ...string) (*Database, error) {
//	dbMu.RLock()
//	defer dbMu.RUnlock()
//
//	if len(Databases) == 0 {
//		return nil, errors.New("no databases available")
//	}
//
//	dbName := "main"
//	if len(name) > 0 {
//		dbName = name[0]
//	}
//
//	db, ok := Databases[dbName]
//	if !ok {
//		return nil, fmt.Errorf("database %s not found", dbName)
//	}
//	return db, nil
//}

// ===== MessagePack 기반 저장/조회 =====
func (d *Database) pack(value any) ([]byte, error) {
	return msgpack.Marshal(value)
}

func (d *Database) unpack(data []byte, dest any) error {
	return msgpack.Unmarshal(data, dest)
}

// Set / Get
func (d *Database) Set(key string, value any, ttl ...time.Duration) error {
	data, err := d.pack(value)
	if err != nil {
		return fmt.Errorf("msgpack marshal failed: %w", err)
	}

	exp := time.Duration(0)
	if len(ttl) > 0 {
		exp = ttl[0]
	}

	return d.client.Set(d.ctx, key, data, exp).Err()
}

func (d *Database) Get(key string, dest any) error {
	data, err := d.client.Get(d.ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return errors.New("key not found")
	}
	if err != nil {
		return fmt.Errorf("redis get error: %w", err)
	}

	return d.unpack(data, dest)
}

// Push / Pop
func (d *Database) LPush(key string, value any) error {
	return d.push("L", key, value)
}

func (d *Database) RPush(key string, value any) error {
	return d.push("R", key, value)
}

func (d *Database) push(direction, key string, value any) error {
	data, err := d.pack(value)
	if err != nil {
		return fmt.Errorf("msgpack marshal failed: %w", err)
	}

	if normalizeDir(direction) == "L" {
		return d.client.LPush(d.ctx, key, data).Err()
	}
	return d.client.RPush(d.ctx, key, data).Err()
}

func (d *Database) LPop(key string, dest any) error {
	return d.pop("L", key, dest)
}

func (d *Database) RPop(key string, dest any) error {
	return d.pop("R", key, dest)
}

func (d *Database) pop(direction, key string, dest any) error {
	var (
		data string
		err  error
	)

	if normalizeDir(direction) == "L" {
		data, err = d.client.LPop(d.ctx, key).Result()
	} else {
		data, err = d.client.RPop(d.ctx, key).Result()
	}

	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("no items in list: %s", key)
	}
	if err != nil {
		return fmt.Errorf("redis pop error: %w", err)
	}

	return d.unpack([]byte(data), dest)
}

// Helpers
func (d *Database) LLen(key string) (int64, error) {
	return d.client.LLen(d.ctx, key).Result()
}

func (d *Database) GetString(key string) (string, error) {
	var v string
	return v, d.Get(key, &v)
}

func (d *Database) GetInt(key string) (int, error) {
	var v int
	return v, d.Get(key, &v)
}

func (d *Database) GetFloat(key string) (float64, error) {
	var v float64
	return v, d.Get(key, &v)
}

func (d *Database) GetBool(key string) (bool, error) {
	var v bool
	return v, d.Get(key, &v)
}

func (d *Database) Delete(key string) error {
	return d.client.Del(d.ctx, key).Err()
}

// Close
func (d *Database) Close() {
	d.client.Close()
}

//func CloseAll() {
//	dbMu.Lock()
//	defer dbMu.Unlock()
//
//	for _, db := range Databases {
//		db.client.Close()
//	}
//	Databases = make(map[string]*Database)
//}

// 내부 함수
func normalizeDir(dir string) string {
	switch dir {
	case "Left", "L", "l":
		return "L"
	default:
		return "R"
	}
}
