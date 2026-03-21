package main

import (
	"fmt"
	"math/rand"
	"time"

	"ratelimiter/limiter"
)

func main() {
	bucket := limiter.NewTokenLimiter(10, 5)
	brustReq := make(chan limiter.Req, 10)
	done := make(chan bool)

	go func() {
		for r := range brustReq {
			bucket.RefilToken(r.Time)
			if bucket.Allowed(r.RequestNo) {
				fmt.Printf("✓ cost=%-2d  tokens remaining: %d \n", r.RequestNo, bucket.GetTokens())
			} else {
				fmt.Printf("✗ cost=%-2d  rejected (only %d tokens)\n", r.RequestNo, bucket.GetTokens())
			}
		}
		close(done)
	}()

	for i := 0; i < 10; i++ {
		req := limiter.Req{
			RequestNo: rand.Intn(10) + 1,
			Time:      time.Now().UnixNano(),
		}
		if i < 9 {
			time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)
		}
		brustReq <- req
	}

	close(brustReq)
	<-done

}
