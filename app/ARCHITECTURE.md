# Cross-Platform Architecture

## Overview

This document explains how the cross-platform architecture works using Go build tags.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         app.go                               │
│                  (Platform-Agnostic Core)                    │
│                                                              │
│  • handleHotkeyPress() - Records when hotkey pressed        │
│  • startRecording()    - Records audio via PortAudio        │
│  • transcribe()        - Transcribes recorded audio         │
│                                                              │
│  Uses abstract functions:                                   │
│  • isHotkeyPressed()           ┐                            │
│  • startHotkeyListener()       │                            │
│  • playRecordingSound()        ├─ Platform implementations  │
│  • checkMicrophonePermission() │                            │
│  • initPlatformSpecific()      ┘                            │
└──────────────────┬──────────────────────┬───────────────────┘
                   │                      │
         ┌─────────▼──────────┐  ┌───────▼────────────┐
         │   macOS Build      │  │   Windows Build    │
         │  (//go:build darwin)│  │ (//go:build windows)│
         └─────────┬──────────┘  └───────┬────────────┘
                   │                      │
    ┌──────────────▼──────────┐  ┌───────▼─────────────┐
    │   cgo_darwin.go         │  │  cgo_windows.go     │
    │   • Fn key (C.fn_pressed)│  │  • F13 key (C code)│
    │   • NSSound             │  │  • MessageBeep      │
    │   • AVFoundation perms  │  │  • No special perms │
    └──────────────┬──────────┘  └───────┬─────────────┘
                   │                      │
    ┌──────────────▼──────────┐  ┌───────▼─────────────┐
    │  platform_darwin.go     │  │ platform_windows.go │
    │  • Check accessibility  │  │ • Start F13 listener│
    │  • Start Fn listener    │  │ • No extra perms    │
    └──────────────┬──────────┘  └───────┬─────────────┘
                   │                      │
    ┌──────────────▼──────────┐  ┌───────▼─────────────┐
    │    C/Objective-C        │  │      C Code         │
    │  • fn_key.c             │  │ • windows_hotkey.c  │
    │  • microphone.m         │  │                     │
    │  • sound.m              │  │                     │
    └─────────────────────────┘  └─────────────────────┘
```

## Build Tag Flow

### On macOS:
```
go build  →  Only compiles:
              • app.go
              • cgo_darwin.go
              • platform_darwin.go
              • fn_key.c, microphone.m, sound.m
              • libportaudio (darwin_arm64 or darwin_amd64)
```

### On Windows:
```
go build  →  Only compiles:
              • app.go
              • cgo_windows.go
              • platform_windows.go
              • windows_hotkey.c
              • libportaudio (windows)
```

## Function Call Chain

### Recording Flow (macOS):

```
User presses Fn key
    ↓
C: fn_key.c detects key press (fn_pressed = 1)
    ↓
Go: isHotkeyPressed() returns C.fn_pressed
    ↓
Go: handleHotkeyPress() checks permission
    ↓
Go: startRecording() called
    ↓
C: PlayStartRecordingSound() plays NSSound
    ↓
Go: PortAudio records from microphone
    ↓
User releases Fn key
    ↓
C: fn_key.c detects release (fn_pressed = 0)
    ↓
Go: stopRecording() called
    ↓
Go: transcribe() processes audio
```

### Recording Flow (Windows):

```
User presses F13 key
    ↓
C: windows_hotkey.c detects key press (hotkey_pressed = 1)
    ↓
Go: isHotkeyPressed() returns C.get_hotkey_state()
    ↓
Go: handleHotkeyPress() checks permission (auto-granted)
    ↓
Go: startRecording() called
    ↓
C: MessageBeep() plays system sound
    ↓
Go: PortAudio records from microphone
    ↓
User releases F13 key
    ↓
C: windows_hotkey.c detects release (hotkey_pressed = 0)
    ↓
Go: stopRecording() called
    ↓
Go: transcribe() processes audio
```

## Permission Handling

### macOS:
```
┌─────────────────────┐
│ First Run           │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐      ┌────────────────┐
│ Check Accessibility │──NO──▶ Show Dialog    │
│ Permission          │      │ Request Access │
└──────────┬──────────┘      └────────────────┘
           │ YES
           ▼
┌─────────────────────┐      ┌────────────────┐
│ Check Microphone    │──NO──▶ Request Access │
│ Permission          │      │ Auto-prompt    │
└──────────┬──────────┘      └────────────────┘
           │ YES
           ▼
┌─────────────────────┐
│ Start Recording     │
└─────────────────────┘
```

### Windows:
```
┌─────────────────────┐
│ First Run           │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Microphone          │
│ Auto-granted ✓      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Start Recording     │
└─────────────────────┘
```

## Key Design Principles

1. **Single Source of Truth**: `app.go` contains all business logic
2. **Platform Isolation**: Platform code only in `cgo_*.go` and `platform_*.go`
3. **Build-Time Selection**: Go build tags ensure only correct code is compiled
4. **Consistent Interface**: All platforms implement the same functions
5. **Zero Runtime Overhead**: No if/else checks for platform at runtime

## Adding a New Platform (e.g., Linux)

1. Create `cgo_linux.go` with `//go:build linux`
2. Implement all required functions:
   - `isHotkeyPressed()`
   - `startHotkeyListener()`
   - `playRecordingSound()`
   - `checkMicrophonePermission()`
   - etc.
3. Create `platform_linux.go` with initialization logic
4. Create C code for hotkey detection (e.g., X11 or Wayland)
5. Add PortAudio library for Linux
6. Build and test!

No changes to `app.go` required!