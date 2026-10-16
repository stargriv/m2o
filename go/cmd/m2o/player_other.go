//go:build !darwin

package main

import "errors"

// player is a placeholder that lets the command be packaged for platforms
// where the macOS AVFoundation backend is not available yet.
type player struct{}

func startPlayer(string) (*player, error) {
	return nil, errors.New("audio playback is currently supported only on macOS")
}

func (p *player) stop() {}
