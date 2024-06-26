package ratelimit

import (
	"golang.org/x/time/rate"
	//"time"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter { // create a new rate limiter
	return &RateLimiter{rate.NewLimiter(r, b)}
}

func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}
