package limiter

import (
	"fmt"
	"time"
)

type Req struct {
	RequestNo int
	Time      int64
}

type TokenLimiter struct {
	capacity  int
	tokens    int
	rate      int
	lastRefil int64
}

func NewTokenLimiter(cap int, rate int) *TokenLimiter {
	return &TokenLimiter{
		capacity:  cap,
		tokens:    cap,
		rate:      rate,
		lastRefil: time.Now().UnixNano(),
	}
}

func (t *TokenLimiter) RefilToken(now int64) {
	elapsedTime := now - t.lastRefil
	tokensAdded := int(float64(elapsedTime) / 1e9 * float64(t.rate))

	if tokensAdded > 0 {
		t.tokens += tokensAdded
		if t.tokens >= t.capacity {
			t.tokens = t.capacity
		}
		t.lastRefil = now
	}

}

func (t *TokenLimiter) Allowed(resquestNo int) bool {
	fmt.Printf("tokens %v : and  costs %v \n ", t.tokens, resquestNo)
	if t.tokens >= resquestNo {
		t.tokens -= resquestNo
		return true
	}
	return false
}

func (t *TokenLimiter) GetTokens() int {
	return t.tokens
}
