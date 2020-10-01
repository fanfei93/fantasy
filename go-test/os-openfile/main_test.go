package main

import (
	"strconv"
	"testing"
)

func BenchmarkMethod1(b *testing.B)  {
	for i := 0; i < b.N; i++ {
		name := strconv.Itoa(i) + ".txt"
		method1(name)
	}
}

func BenchmarkMethod2(b *testing.B)  {
	for i := 0; i < b.N; i++ {
		name := strconv.Itoa(i) + ".txt"
		method2(name)
	}
}