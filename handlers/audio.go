package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/generative-ai-go/genai"
	"gofr.dev/pkg/gofr"
	"google.golang.org/api/option"
)

type AudioRequest struct {
	Audio *multipart.FileHeader `file:"audio"`
}

type AudioResponse struct {
	Reply     string `json:"reply"`
	AudioPath string `json:"audioPath"`
}

type ElevenLabsRequest struct {
	Text          string                 `json:"text"`
	VoiceSettings map[string]interface{} `json:"voice_settings"`
	Language      string                 `json:"language"`
}

func AudioChatHandler(ctx *gofr.Context) (interface{}, error) {
	var req AudioRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("no audio uploaded")
	}
	if req.Audio == nil {
		return nil, fmt.Errorf("no audio uploaded")
	}
	file, err := req.Audio.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file")
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "audio_*"+filepath.Ext(req.Audio.Filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file")
	}
	defer os.Remove(tempFile.Name())

	if _, err := io.Copy(tempFile, file); err != nil {
		return nil, fmt.Errorf("failed to copy file")
	}
	tempFile.Close()

	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.CloudAPIKey, cfg.CloudAPISecret)
	if err == nil {
		_, _ = cld.Upload.Upload(context.Background(), tempFile.Name(), uploader.UploadParams{ResourceType: "video", Folder: "audio_files"})
	}

	genaiClient, err := genai.NewClient(context.Background(), option.WithAPIKey(firstNonEmpty(cfg.ChatGeminiKey, cfg.GeminiKey)))
	if err != nil {
		return nil, fmt.Errorf("failed to create GenAI client")
	}
	defer genaiClient.Close()

	fileData, _ := os.ReadFile(tempFile.Name())
	uploadedFile, err := genaiClient.UploadFile(context.Background(), "", bytes.NewReader(fileData), &genai.UploadFileOptions{MIMEType: req.Audio.Header.Get("Content-Type")})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to GenAI")
	}

	model := genaiClient.GenerativeModel("gemini-1.5-pro")
	prompt := `You are an assistant AI.
1. Transcribe accurately.
2. Detect language.
3. Answer in same language.
4. Append [xx] language code.`
	resp, err := model.GenerateContent(context.Background(), genai.FileData{URI: uploadedFile.URI, MIMEType: uploadedFile.MIMEType}, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content")
	}

	var replyText string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			replyText = string(textPart)
		}
	}
	if replyText == "" {
		replyText = "No response from Gemini."
	}

	langCodeRegex := regexp.MustCompile(`\[([a-z]{2})\]$`)
	matches := langCodeRegex.FindStringSubmatch(replyText)
	langCode := "en"
	if len(matches) > 1 {
		langCode = matches[1]
	}
	replyText = strings.TrimSpace(langCodeRegex.ReplaceAllString(replyText, ""))

	speechBuffer, err := generateSpeechWithElevenLabs(replyText, langCode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate speech")
	}
	if err := os.MkdirAll("public/audio", 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio directory")
	}
	audioFileName := fmt.Sprintf("reply_%d.mp3", time.Now().UnixNano())
	audioPath := filepath.Join("public/audio", audioFileName)
	if err := os.WriteFile(audioPath, speechBuffer, 0644); err != nil {
		return nil, fmt.Errorf("failed to save audio file")
	}
	return AudioResponse{Reply: replyText, AudioPath: "/audio/" + audioFileName}, nil
}

func generateSpeechWithElevenLabs(text, langCode string) ([]byte, error) {
	apiKey := firstNonEmpty(cfg.ChatElevenKey, cfg.ElevenLabsKey)
	voiceID := "iWNf11sz1GrUE4ppxTOL"
	if apiKey == "" {
		return nil, fmt.Errorf("ElevenLabs API key not configured")
	}
	reqBody := ElevenLabsRequest{Text: text, VoiceSettings: map[string]interface{}{"stability": 0.5, "similarity_boost": 0.75}, Language: langCode}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API error: %s", string(b))
	}
	return io.ReadAll(resp.Body)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
