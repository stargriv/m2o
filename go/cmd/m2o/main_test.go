package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrackName(t *testing.T) {
	if got, want := trackName("Take On Me A-HA"), "A-Ha - Take On Me"; got != want {
		t.Fatalf("trackName() = %q, want %q", got, want)
	}
}

func TestGetOnAir(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"audio_url":"https://audio.example/stream","title":"Song ARTIST"}`))
	}))
	defer server.Close()

	got, err := getOnAir(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got.AudioURL != "https://audio.example/stream" || got.Title != "Song ARTIST" {
		t.Fatalf("unexpected response: %#v", got)
	}
}

func TestResolveStreamFollowsRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/live", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("audio"))
	}))
	defer server.Close()

	got, err := resolveStream(context.Background(), server.URL+"/start")
	if err != nil {
		t.Fatal(err)
	}
	if got != server.URL+"/live" {
		t.Fatalf("resolveStream() = %q, want %q", got, server.URL+"/live")
	}
}
