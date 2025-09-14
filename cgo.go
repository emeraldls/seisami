package main

/*
#cgo CFLAGS: -I./include
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include "fn_key.c"

#cgo CFLAGS: -I./include -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#include "sound.m"

#cgo CFLAGS: -I./include -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AVFoundation -framework AppKit
#include "microphone.m"

#cgo darwin,arm64 CFLAGS: -I${SRCDIR}/portaudio/include
#cgo darwin,arm64 LDFLAGS: ${SRCDIR}/portaudio/lib/darwin_arm64/libportaudio.a -framework CoreAudio -framework AudioToolbox -framework AudioUnit -framework CoreServices -framework Carbon

#cgo darwin,amd64 CFLAGS: -I${SRCDIR}/portaudio/include
#cgo darwin,amd64 LDFLAGS: ${SRCDIR}/portaudio/lib/darwin_amd64/libportaudio.a -framework CoreAudio -framework AudioToolbox -framework AudioUnit -framework CoreServices -framework Carbon
#include "portaudio.h"
*/
import "C"

func isFnPressed() bool {
	return C.fn_pressed != 0
}

func startListener() {
	C.start_listener()
}

func PlaySound() {
	C.PlayStartRecordingSound()
}

func checkMicrophonePermission() int {
	return int(C.check_microphone_permission())
}

func requestMicrophonePermission() {
	C.request_microphone_permission()
}

func isPermissionRequestCompleted() bool {
	return C.is_permission_request_completed() != 0
}

func getPermissionRequestResult() int {
	return int(C.get_permission_request_result())
}

func requestMicrophonePermissionSync() bool {
	return C.request_microphone_permission_sync() != 0
}

func openMicrophoneSettings() {
	C.open_microphone_settings()
}
