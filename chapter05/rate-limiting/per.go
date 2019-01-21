package main

import "time"
import "golang.org/x/time/rate"

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}
