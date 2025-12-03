# Redis Store Wrapper (Improved)


Redis에 데이터를 **messagePack 직렬화**하여 저장하고, 다양한 타입으로 안전하게 가져올 수 있도록 개선된 래퍼입니다.  
기존의 단순 문자열 저장 방식을 넘어, 구조체와 같은 복잡한 데이터 타입도 쉽게 다룰 수 있습니다.
---

## ✨ 주요 기능
- `New` : Redis 클라이언트 인스턴스 생성
- `Set` : 값을 messagePack으로 직렬화하여 저장
- `Get` : Redis에서 messagePack 데이터를 가져와 원하는 구조체로 언마샬링
- `LPush, RPush` : 리스트에 값 추가
- `LPop, RPop` : 리스트에서 값 제거 및 반환
- `LLen` : 리스트 길이 조회
- `Delete` : 키 삭제
- `Reset` : 전체 데이터 삭제 (확인 플래그 필요)
- `Close` : 연결 종료

---

## 📦 설치
```bash
go get github.com/pro200/go-redis
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
	Name string `msgpack:"name"`
	Age  int    `msgpack:"age"`
}

func main() {
	// Redis 클라이언트 생성 (기본 설정)
	rds := redis.New()
	defer rds.Close()

	// 사용자 정의 설정
	/*
	rds := redis.New(redis.Config{
		Host:     "localhost",
		Port:     6380,
		Password: "mypassword",
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
	TLSConfig: nil,
	PoolSize:  10 * runtime.GOMAXPROCS(0),
}
```

