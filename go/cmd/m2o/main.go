// m2o-player plays Radio M2O and prints changes to the currently playing track.
//
// On macOS, playback is handled by the built-in AVFoundation framework. This
// keeps the executable self-contained apart from macOS system frameworks.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	onAirURL          = "https://www.m2o.it/api/pub/v2/all/gdwc-audio-player/onair?format=json"
	trackPollInterval = 20 * time.Second
)

type onAirResponse struct {
	AudioURL string `json:"audio_url"`
	Title    string `json:"title"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func getOnAir(ctx context.Context, endpoint string) (onAirResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return onAirResponse{}, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return onAirResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return onAirResponse{}, fmt.Errorf("GET %s: %s", endpoint, resp.Status)
	}
	var result onAirResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return onAirResponse{}, err
	}
	if result.AudioURL == "" {
		return onAirResponse{}, errors.New("on-air response has no audio_url")
	}
	return result, nil
}

// resolveStream follows the redirect supplied by the API without downloading
// the audio body. AVFoundation is given this final URL.
func resolveStream(ctx context.Context, streamURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("GET stream: %s", resp.Status)
	}
	return resp.Request.URL.String(), nil
}

func trackName(title string) string {
	words := strings.Fields(title)
	artistStart := len(words)
	for i := len(words) - 1; i >= 0; i-- {
		word := words[i]
		if isUppercaseArtistWord(word) || (word == "&" && artistStart < len(words)) {
			artistStart = i
			continue
		}
		break
	}
	if artistStart == len(words) {
		return title
	}
	return strings.Title(strings.ToLower(strings.Join(words[artistStart:], " "))) + " - " + strings.Join(words[:artistStart], " ")
}

func isUppercaseArtistWord(word string) bool {
	hasLetter := false
	for _, r := range word {
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
			continue
		}
		if r >= 'a' && r <= 'z' {
			return false
		}
	}
	return hasLetter
}

func monitorTrack(ctx context.Context, endpoint string, out io.Writer) {
	var previous string
	ticker := time.NewTicker(trackPollInterval)
	defer ticker.Stop()
	for {
		onAir, err := getOnAir(ctx, endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not fetch track information: %v\n", err)
		} else if track := trackName(onAir.Title); track != previous {
			fmt.Fprintf(out, "%s %s\n", time.Now().Format("15:04 02.01.2006"), track)
			previous = track
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func run(ctx context.Context, endpoint string) error {
	onAir, err := getOnAir(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("get on-air information: %w", err)
	}
	stream, err := resolveStream(ctx, onAir.AudioURL)
	if err != nil {
		return fmt.Errorf("resolve audio stream: %w", err)
	}
	player, err := startPlayer(stream)
	if err != nil {
		return fmt.Errorf("start audio playback: %w", err)
	}
	defer player.stop()
	go monitorTrack(ctx, endpoint, os.Stdout)
	<-ctx.Done()
	return nil
}

func main() {
	// AVFoundation/AppKit must be initialized and used on one OS main thread.
	runtime.LockOSThread()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, onAirURL); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
