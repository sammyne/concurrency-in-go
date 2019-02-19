package main

import (
	"context"
	"sort"

	"golang.org/x/time/rate"
)

type RateLimiter interface { // <1>
	Wait(context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit) // <2>
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil

	//return l.limiters[0].Wait(ctx)
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit() // <3>: the most restrictive limit due to sorting
}
