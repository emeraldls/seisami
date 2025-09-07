package main

/*
#cgo CFLAGS: -I./include -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#include "sound.m"
*/
import "C"

func PlaySound() {

	C.PlayStartRecordingSound()
}
