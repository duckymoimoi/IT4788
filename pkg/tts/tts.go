package tts

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// VoiceEntry defines a voice instruction audio file.
type VoiceEntry struct {
	Key  string // e.g. "turn_left"
	Text string // Vietnamese text to synthesize
}

// DefaultVoices contains the 8 navigation voice instructions.
var DefaultVoices = []VoiceEntry{
	{Key: "turn_left", Text: "Rẽ trái"},
	{Key: "turn_right", Text: "Rẽ phải"},
	{Key: "go_straight", Text: "Đi thẳng"},
	{Key: "arrived", Text: "Bạn đã đến đích"},
	{Key: "elevator_up", Text: "Đi thang máy lên"},
	{Key: "elevator_down", Text: "Đi thang máy xuống"},
	{Key: "stairs_up", Text: "Đi cầu thang lên"},
	{Key: "stairs_down", Text: "Đi cầu thang xuống"},
}

// GenerateAll downloads all voice files from Google Translate TTS.
// Files are saved to the specified directory.
// Skips files that already exist to avoid re-downloading.
func GenerateAll(audioDir string) error {
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return fmt.Errorf("cannot create audio dir: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	generated := 0

	for _, v := range DefaultVoices {
		filePath := filepath.Join(audioDir, v.Key+".mp3")

		// Skip if already exists
		if _, err := os.Stat(filePath); err == nil {
			continue
		}

		// Google Translate TTS endpoint (free, no API key needed)
		ttsURL := fmt.Sprintf(
			"https://translate.google.com/translate_tts?ie=UTF-8&client=tw-ob&q=%s&tl=vi",
			url.QueryEscape(v.Text),
		)

		resp, err := client.Get(ttsURL)
		if err != nil {
			log.Printf("[TTS] WARN: Cannot download '%s': %v", v.Key, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Printf("[TTS] WARN: HTTP %d for '%s'", resp.StatusCode, v.Key)
			continue
		}

		out, err := os.Create(filePath)
		if err != nil {
			resp.Body.Close()
			log.Printf("[TTS] WARN: Cannot create file '%s': %v", filePath, err)
			continue
		}

		written, err := io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()

		if err != nil {
			os.Remove(filePath) // Clean up partial file
			log.Printf("[TTS] WARN: Failed writing '%s': %v", v.Key, err)
			continue
		}

		generated++
		log.Printf("[TTS] Generated %s.mp3 (%d bytes)", v.Key, written)

		// Small delay to be polite to Google
		time.Sleep(300 * time.Millisecond)
	}

	if generated > 0 {
		log.Printf("[TTS] Generated %d new voice files in %s", generated, audioDir)
	} else {
		log.Printf("[TTS] All %d voice files already exist in %s", len(DefaultVoices), audioDir)
	}

	return nil
}
