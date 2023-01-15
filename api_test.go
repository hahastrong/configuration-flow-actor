package cfa

import (
	"fmt"
	"testing"
	"time"
)

func TestApi(t *testing.T) {
	start := time.Now().UnixMilli()
	request := `{"url":"https://www.youtube.com/watch?v=q12vwFMgqfk","playlist":true,"format":"mkv"}`
	rsp, err := CFARun(nil, "helo", request)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(time.Now().UnixMilli() - start, string(rsp))
	time.Sleep(time.Second * 100)
}
