package ratelimit

import (
	"golang.org/x/time/rate"
	//"time"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter { // create a new rate limiter
}

func (rl *RateLimiter) Allow() bool {
	// check if the action is allowed by the rate limiter
}
