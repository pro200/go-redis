package redis_test

import (
	"testing"
	"time"

	"github.com/pro200/go-redis"
)

func TestRedis(t *testing.T) {
	// redis.io: database-dev for dev
	rds := redis.New(redis.Config{
		Host:     "redis-15029.c340.ap-northeast-2-1.ec2.redns.redis-cloud.com",
		Port:     15029,
		Username: "default",
		Password: "OuwZQqjmrZMcuIeQtb97E3ATUIXevsEt",
	})
	defer rds.Close()

	// 값 저장
	err := rds.Set("test", struct {
		Name string
		Age  int
	}{
		Name: "Alice",
		Age:  18,
	}, 10*time.Minute)

	if err != nil {
		t.Error("Set error:", err)
	}

	// 구조체로 값 조회
	var user struct {
		Name string
		Age  int
	}

	if err := rds.Get("test", &user); err != nil {
		t.Error("Get error:", err)
	}

	if user.Name != "Alice" {
		t.Error("Wrong result")
	}

	// 키 삭제
	err = rds.Delete("test")
	if err != nil {
		t.Error("Delete error:", err)
	}
}
