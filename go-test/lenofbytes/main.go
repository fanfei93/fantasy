package main

var d int = 9
var e byte = 1 << 9 / 128


var i int8 = 64
var f int8 = i * i / 64

func main() {
	println(e)
	println(f)
}