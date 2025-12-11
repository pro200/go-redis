package redis_test

import (
	"testing"
	"time"

	"github.com/pro200/go-redis"
)

type User struct {
	Name string `msgpack:"name"`
	Age  int    `msgpack:"age"`
}

func TestRedis(t *testing.T) {
	rds, err := redis.NewDatabase()
	if err != nil {
		t.Error("NewDatabase error:", err)
	}
	defer rds.Close()

	// 값 저장
	err = rds.Set("test", User{
		Name: "Alice",
		Age:  30,
	}, time.Minute*5)

	if err != nil {
		t.Error("Set error:", err)
	}

	// 구조체로 값 조회
	var user User

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

	// 리스트에 값 추가 및 조회
	var value = map[string]string{
		"url":     "pro200.diskn.com/012345/abcdefg",
		"size":    "1234",
		"changed": "3",
	}

	err = rds.RPush("test_list", value)
	if err != nil {
		t.Error("Push error:", err)
	}

	var result map[string]string
	err = rds.LPop("test_list", &result)
	if err != nil {
		t.Error("Pop error:", err)
	}

	if result["url"] != "pro200.diskn.com/012345/abcdefg" {
		t.Error("Wrong list result")
	}
}
