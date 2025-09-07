package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func TranscribeElevenLabs(fileURL string) (string, error) {
	if cfg.ElevenLabsKey == "" {
		return "", fmt.Errorf("ElevenLabs API key not configured")
	}
	body := strings.NewReader(fmt.Sprintf(`{"url": "%s", "language": "en"}`, fileURL))
	req, err := http.NewRequest("POST", "https://api.elevenlabs.io/v1/speech-to-text", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", cfg.ElevenLabsKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ElevenLabs API error: %s", resp.Status)
	}
	var result struct{ Text, Error string }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf(result.Error)
	}
	return result.Text, nil
}

func GenerateHashtags(ctx context.Context, text string) ([]string, error) {
	if cfg.GeminiKey == "" {
		return nil, fmt.Errorf("gemini API key not configured")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := fmt.Sprintf("Generate many relevant hashtags in English, only hashtags separated by spaces. Text:\n\n%s", text)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, c := range resp.Candidates {
		if c.Content != nil {
			for _, part := range c.Content.Parts {
				if t, ok := part.(genai.Text); ok {
					for _, w := range strings.Fields(string(t)) {
						if strings.HasPrefix(w, "#") {
							tags = append(tags, w)
						}
					}
				}
			}
		}
	}
	return tags, nil
}
