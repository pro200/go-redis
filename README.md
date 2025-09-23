# Redis Store Wrapper (Improved)

이 패키지는 [Fiber Redis Storage](https://github.com/gofiber/storage/tree/main/redis) 기반으로  
Redis에 데이터를 **JSON 직렬화**하여 저장하고, 다양한 타입으로 안전하게 가져올 수 있도록 개선된 래퍼입니다.  
전역 인스턴스 대신 구조체(`Store`)를 사용해 멀티 인스턴스 환경에서도 안정적으로 동작합니다.

---

## ✨ 주요 기능
- `New` : Redis 클라이언트 인스턴스 생성
- `Set` : 값을 JSON으로 직렬화하여 저장
- `Get` : Redis에서 JSON 데이터를 가져와 원하는 구조체로 언마샬링
- `Delete` : 키 삭제
- `Reset` : 전체 데이터 삭제 (확인 플래그 필요)
- `Close` : 연결 종료

---

## 📦 설치
```bash
go get github.com/gofiber/storage/redis/v3
```

## 🚀 사용 예시
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pro200/go-redis" // 모듈 경로에 맞게 수정
)

type User struct {
	Name string
	Age  int
}

func main() {
	// Redis 클라이언트 생성 (기본 설정)
	store := redis.New()
	defer store.Close()

	// 사용자 정의 설정
	/*
	rds := redis.New(redis.Config{
		Host:     "localhost",
		Port:     6380,
		Password: "mypassword",
		Database: 1,
	})
	*/

	// 값 저장
	err := rds.Set("user:1", User{
		"name": "Alice",
		"age":  30,
	}, 10*time.Minute)
	if err != nil {
		log.Fatal("Set error:", err)
	}

	// 구조체로 값 조회
	var user User
	if err := rds.Get("user:1", &user); err != nil {
		log.Fatal("Get error:", err)
	}
	fmt.Println("User:", user)

	// 키 삭제
	_ = rds.Delete("user:1")

	// 전체 삭제 (안전장치: true 전달)
	// _ = store.Reset(true)
}
```

## ⚙️ 기본 Config
```go
redis.Config{
	Host:      "127.0.0.1",
	Port:      6379,
	Database:  0,
	Username:  "",
	Password:  "",
	Reset:     false,
	TLSConfig: nil,
	PoolSize:  10 * runtime.GOMAXPROCS(0),
}
```

