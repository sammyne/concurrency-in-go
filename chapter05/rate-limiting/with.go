// +build with
package main

import (
	"context"

	"golang.org/x/time/rate"
)

func Open() *APIConnection {
	return &APIConnection{
		rateLimiter: rate.NewLimiter(rate.Limit(1), 1), // <1>
	}
}

type APIConnection struct {
	rateLimiter *rate.Limiter
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil { // <2>
		return err
	}
	// Pretend we do work here
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil { // <2>
		return err
	}
	// Pretend we do work here
	return nil
}
