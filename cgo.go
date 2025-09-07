package main

/*
#cgo CFLAGS: -I./include
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include "fn_key.c"
*/
import "C"

func isFnPressed() bool {
	return C.fn_pressed != 0
}

func startListener() {
	C.start_listener()
}
