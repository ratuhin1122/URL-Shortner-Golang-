package routes

import (
	"time"

)
type Rquest struct{
	URL           string        `json:"url"`
	CustomShort   string        `json:"custom_short"`
	Expiry        time.Duration `json:"expiry"`
}

type Response struct{
	URL                 string        `json:"url"`
	CustomShort         string        `json:"custom_short"`
	Expiry              time.Duration `json:"expiry"`
	XRateRemaining      int           `json:"rate_limit_remaining"`
	XRateLimitReset     int64         `json:"rate_limit_reset"`
}