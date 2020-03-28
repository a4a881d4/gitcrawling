package types

import "fmt"

type IntCounter struct {
	Len   int
	C     []int64
	Total int64
}

func NewIntCounter(l int) *IntCounter {
	return &IntCounter{
		Len: l,
		C:   make([]int64, l),
	}
}
func (ic *IntCounter) Count32(a uint32) int {
	return ic.Count64(int64(a))
}
func (ic *IntCounter) Count64(a int64) int {
	var r int = 0
	for ; a != 0 && r < ic.Len; r++ {
		a = a >> 1
	}
	ic.C[r]++
	ic.Total++
	return r
}
func (ic *IntCounter) Count64Other(a, b int64) int {
	var r int = 0
	for ; a != 0 && r < ic.Len; r++ {
		a = a >> 1
	}
	ic.C[r] += b
	ic.Total += b
	return r
}
func (ic *IntCounter) Dump() {
	if ic.Total == 0 {
		fmt.Println("nil Counter")
		return
	}
	for i, v := range ic.C {
		if i == 0 {
			fmt.Printf("%10d: %8d ", 0, v*1_000_000/ic.Total)
		} else {
			fmt.Printf("%10d: %8d ", 1<<(i-1), v*1_000_000/ic.Total)
		}
		if i&3 == 3 {
			fmt.Println()
		}
	}
}
