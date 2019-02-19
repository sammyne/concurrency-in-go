// +build without

package main

//go:generate go run main.go without.go

import "context"

func Open() *APIConnection {
	return &APIConnection{}
}

type APIConnection struct{}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	// Pretend we do work here
	return nil
}
func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	// Pretend we do work here
	return nil
}
