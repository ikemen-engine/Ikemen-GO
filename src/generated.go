package main

func AppendI(slice []int, data ...int) []int {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]int, n+n/4)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[:n]
	copy(slice[m:n], data)
	return slice
}
func AppendU32(slice []uint32, data ...uint32) []uint32 {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]uint32, n+n/4)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[:n]
	copy(slice[m:n], data)
	return slice
}
func AppendPal(slice [][]uint32, data ...[]uint32) [][]uint32 {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([][]uint32, n+n/4)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[:n]
	copy(slice[m:n], data)
	return slice
}
func AppendAF(slice []AnimFrame, data ...AnimFrame) []AnimFrame {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]AnimFrame, n+n/4)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[:n]
	copy(slice[m:n], data)
	return slice
}
