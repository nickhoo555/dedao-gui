package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAudioCommandsDiscoverFFmpegWhenSettingEmpty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX executable fixture")
	}
	dir := t.TempDir()
	marker := filepath.Join(dir, "called")
	if err := os.WriteFile(filepath.Join(dir, "ffmpeg"), []byte("#!/bin/sh\nprintf called > \"$FFMPEG_TEST_MARKER\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("FFMPEG_TEST_MARKER", marker)
	old := FfmpegDir
	FfmpegDir = ""
	t.Cleanup(func() { FfmpegDir = old })
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"audio", func() error { return MergeAudio(nil, filepath.Join(dir, "audio.mp3")) }},
		{"audio-video", func() error { return MergeAudioAndVideo(nil, filepath.Join(dir, "video.mp4")) }},
		{"mp4", func() error { return MergeToMP4(nil, filepath.Join(dir, "merged.mp4"), filepath.Join(dir, "parts")) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); err != nil {
				t.Fatalf("empty setting must discover ffmpeg: %v", err)
			}
			if _, err := os.Stat(marker); err != nil {
				t.Fatalf("ffmpeg not invoked: %v", err)
			}
			if err := os.Remove(marker); err != nil {
				t.Fatal(err)
			}
		})
	}
}
