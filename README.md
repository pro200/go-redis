```go
package main

import (
    "fmt"
	"github.com/pro200/go-redis"
)

// 예제 데이터 구조체
type Example struct {
    Name  string
    Age   int
    Email string
}

func main() {
    // Redis 클라이언트
    redis.Init()
    defer redis.Close()

	// 데이터 생성
	data := Example{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}

	// Redis에 데이터 저장
	err := redis.Set("user:1001", data, time.Second*10)
	if err != nil {
		fmt.Println("Failed to set value:", err)
	}

	// Redis에서 언마샬링된 데이터 가져오기
	var rdata Example
	err = redis.Get("user:1001", &rdata)
	if err != nil {
		fmt.Println("Failed to get value:", err)
		return
	}

	// 가져온 데이터 출력
	fmt.Printf("Retrieved Data: %+v\n", rdata)
}
```