package redis

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type Config struct {
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

func NewDatabase(cfg ...Config) *Database {
	c := Config{
		Host:     "127.0.0.1",
		Port:     6379,
		Database: 0,
	}

	if len(cfg) > 0 {
		tmp := cfg[0]
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

	return &Database{
		client: redis.NewClient(options),
		ctx:    context.Background(),
	}
}

// Set & Get
func (d *Database) Set(key string, value any, ttl ...time.Duration) error {
	data, err := pack(value)
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

	return unpack(data, dest)
}

// Push / Pop
func (d *Database) LPush(key string, value any) error {
	return push(d, "L", key, value)
}

func (d *Database) RPush(key string, value any) error {
	return push(d, "R", key, value)
}

func push(db *Database, direction, key string, value any) error {
	data, err := pack(value)
	if err != nil {
		return fmt.Errorf("msgpack marshal failed: %w", err)
	}

	if normalizeDir(direction) == "L" {
		return db.client.LPush(db.ctx, key, data).Err()
	}
	return db.client.RPush(db.ctx, key, data).Err()
}

func (d *Database) LPop(key string, dest any) error {
	return pop(d, "L", key, dest)
}

func (d *Database) RPop(key string, dest any) error {
	return pop(d, "R", key, dest)
}

func pop(db *Database, direction, key string, dest any) error {
	var (
		data []byte
		err  error
	)

	if normalizeDir(direction) == "L" {
		data, err = db.client.LPop(db.ctx, key).Bytes()
	} else {
		data, err = db.client.RPop(db.ctx, key).Bytes()
	}

	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("no items in list: %s", key)
	}
	if err != nil {
		return fmt.Errorf("redis pop error: %w", err)
	}

	return unpack(data, dest)
}

func (d *Database) LTrim(key string, start, stop int64) error {
	return d.client.LTrim(d.ctx, key, start, stop).Err()
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

/*
 * 내부 함수
 */
func pack(value any) ([]byte, error) {
	return msgpack.Marshal(value)
}

func unpack(data []byte, dest any) error {
	return msgpack.Unmarshal(data, dest)
}

func normalizeDir(dir string) string {
	switch dir {
	case "Left", "L", "l":
		return "L"
	default:
		return "R"
	}
}
