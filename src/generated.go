package main

import "image/color"

func AppendI(slice *[]int, data ...int) {
	m := len(*slice)
	n := m + len(data)
	if n > cap(*slice) {
		newSlice := make([]int, n+n/4)
		copy(newSlice, *slice)
		*slice = newSlice
	}
	*slice = (*slice)[:n]
	copy((*slice)[m:n], data)
}
func AppendPal(slice *[][]color.Color, data ...[]color.Color) {
	m := len(*slice)
	n := m + len(data)
	if n > cap(*slice) {
		newSlice := make([][]color.Color, n+n/4)
		copy(newSlice, *slice)
		*slice = newSlice
	}
	*slice = (*slice)[:n]
	copy((*slice)[m:n], data)
}
