package redis_test

import (
	"testing"

	"github.com/pro200/go-redis"
)

func TestRedis(t *testing.T) {
	// redis.io: database-dev for dev
	redis.Init(redis.Config{
		Host:     "redis-15029.c340.ap-northeast-2-1.ec2.redns.redis-cloud.com",
		Port:     15029,
		Username: "default",
		Password: "OuwZQqjmrZMcuIeQtb97E3ATUIXevsEt",
	})
	defer redis.Close()

	if err := redis.Set("test", "hello"); err != nil {
		t.Error(err)
	}

	var result string
	if err := redis.Get("test", &result); err != nil {
		t.Error(err)
	}
	if result != "hello" {
		t.Error("Wrong result")
	}

	if err := redis.Delete("test"); err != nil {
		t.Error(err)
	}
}
