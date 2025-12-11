package redis

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
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
	version float64
	client  *redis.Client
	ctx     context.Context
}

func NewDatabase(cfg ...Config) (*Database, error) {
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

	database := &Database{
		client: redis.NewClient(options),
		ctx:    context.Background(),
	}

	// 버전 확인
	info, err := database.client.Info(database.ctx, "server").Result()
	if err != nil {
		return nil, fmt.Errorf("redis info fetch failed: %v", err)
	}

	re := regexp.MustCompile(`redis_version:(\d+)\.(\d+)`)
	m := re.FindStringSubmatch(info)
	if len(m) < 3 {
		return nil, fmt.Errorf("redis_version not found in INFO")
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	database.version, _ = strconv.ParseFloat(fmt.Sprintf("%d.%d", major, minor), 64)

	return database, nil
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

func (d *Database) LPopCount(key string, count int, dest any) error {
	return popCount(d, "L", key, count, dest)
}

func (d *Database) RPopCount(key string, count int, dest any) error {
	return popCount(d, "R", key, count, dest)
}

func popCount(db *Database, direction, key string, count int, dest any) error {
	if count <= 0 {
		return fmt.Errorf("count must be positive")
	}

	var (
		data []string
		err  error
	)

	if db.version >= 6.2 {
		if normalizeDir(direction) == "L" {
			data, err = db.client.LPopCount(db.ctx, key, count).Result()
		} else {
			data, err = db.client.RPopCount(db.ctx, key, count).Result()
		}

		if errors.Is(err, redis.Nil) || len(data) == 0 {
			return fmt.Errorf("no items in list: %s", key)
		}
		if err != nil {
			return fmt.Errorf("redis pop count error: %w", err)
		}
	} else {
		// Redis 6.2 미만에서는 여러 개 팝을 지원하지 않음
		pipe := db.client.TxPipeline()

		// 0 ~ count-1 까지 가져오기
		r1 := pipe.LRange(db.ctx, key, 0, int64(count-1))
		// 앞에서 count 개 잘라내기
		pipe.LTrim(db.ctx, key, int64(count), -1)

		_, err := pipe.Exec(db.ctx)
		if err != nil && !errors.Is(err, redis.Nil) {
			return fmt.Errorf("redis transaction error: %w", err)
		}

		items, err := r1.Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return fmt.Errorf("lrange error: %w", err)
		}

		if len(items) == 0 {
			return fmt.Errorf("no items in list: %s", key)
		}

		data = items
	}

	/*
	 * Decode
	 */
	// dest는 *[]T 이어야 함
	// data = []string → msgpack 여러 개 언패킹
	slice := make([]any, 0, len(data))

	for _, s := range data {
		var item any
		// string → []byte
		if err := msgpack.Unmarshal([]byte(s), &item); err != nil {
			return fmt.Errorf("msgpack unmarshal failed: %w", err)
		}
		slice = append(slice, item)
	}

	// []any → 최종 dest 타입으로 언패킹
	packed, err := msgpack.Marshal(slice)
	if err != nil {
		return fmt.Errorf("slice repack failed: %w", err)
	}

	return msgpack.Unmarshal(packed, dest)
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

func (d *Database) GetInt(key string) (int64, error) {
	var v int64
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
