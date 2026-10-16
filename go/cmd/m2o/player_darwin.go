package main

/*
#cgo LDFLAGS: -framework AVFoundation -framework Foundation -framework AppKit
#include <stdlib.h>

void *M2OPlayerStart(const char *streamURL, char **errorMessage);
void M2OPlayerStop(void *player);
*/
import "C"

import (
	"errors"
	"unsafe"
)

// player is an AVFoundation AVPlayer retained by the Objective-C bridge.
type player struct{ handle unsafe.Pointer }

func startPlayer(streamURL string) (*player, error) {
	cURL := C.CString(streamURL)
	defer C.free(unsafe.Pointer(cURL))

	var cError *C.char
	handle := C.M2OPlayerStart(cURL, &cError)
	if cError != nil {
		defer C.free(unsafe.Pointer(cError))
		return nil, errors.New(C.GoString(cError))
	}
	if handle == nil {
		return nil, errors.New("AVFoundation did not create a player")
	}
	return &player{handle: handle}, nil
}

func (p *player) stop() {
	if p != nil && p.handle != nil {
		C.M2OPlayerStop(p.handle)
		p.handle = nil
	}
}
