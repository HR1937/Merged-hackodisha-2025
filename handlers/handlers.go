package handlers

import (
	"fmt"
)

type ServerConfig struct {
	JWTSecret          string
	ElevenLabsKey      string
	GeminiKey          string
	ChatGeminiKey      string
	ChatElevenKey      string
	NeighbourGeminiKey string
	NeighbourElevenKey string
	CloudName          string
	CloudAPIKey        string
	CloudAPISecret     string
	FirestoreProjectID string
	GoogleCredentials  string
}

var cfg ServerConfig

func SetConfig(c ServerConfig) {
	cfg = c
	fmt.Println("handler config initialized")
}
