package main

import "math"

func Abs(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}
func IsFinite(f float32) bool {
	return math.Abs(float64(f)) <= math.MaxFloat64
}

type Error string

func (e Error) Error() string { return string(e) }
