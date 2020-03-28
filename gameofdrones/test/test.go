package main

import (
	"fmt"
	"math"
)

type S struct {
	x, y int
}

type SortWrapper struct {
	s  []*S
	by func(p, q *S) bool
}

func (sw SortWrapper) Len() int { // 重写 Len() 方法
	return len(sw.s)
}
func (sw SortWrapper) Swap(i, j int) { // 重写 Swap() 方法
	sw.s[i], sw.s[j] = sw.s[j], sw.s[i]
}
func (sw SortWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return sw.by(sw.s[i], sw.s[j])
}

func main() {

	norm := math.Sqrt(math.Pow(float64(10), 2) + math.Pow(float64(10), 2))
	fmt.Println(norm)

	a := 1
	x := int(float64(a) / 11)
	fmt.Println(x)
}
