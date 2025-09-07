package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	"finalapp/handlers"
	"finalapp/store"

	"gofr.dev/pkg/gofr"
)

type Config struct {
	MongoURI           string `json:"mongo_uri"`
	DBName             string `json:"db_name"`
	JWTSecret          string `json:"jwt_secret"`
	Port               string `json:"port"`
	GeminiKeyBase      string `json:"gemini_api_key"`
	ElevenKeyBase      string `json:"elevenlabs_api_key"`
	GeminiKeyChat      string `json:"chatbot_gemini_api_key"`
	ElevenKeyChat      string `json:"chatbot_eleven_api_key"`
	GeminiKeyNeighbour string `json:"neighbour_gemini_api_key"`
	ElevenKeyNeighbour string `json:"neighbour_eleven_api_key"`
	FirestoreProjectID string `json:"firestore_project_id"`
	GoogleCredentials  string `json:"google_application_credentials"`
	CloudName          string `json:"cloudinary_cloud"`
	CloudAPIKey        string `json:"cloudinary_api_key"`
	CloudAPISecret     string `json:"cloudinary_api_secret"`
}

func pickFreePort(candidates []string, fallback string) string {
	for _, p := range candidates {
		ln, err := net.Listen("tcp", ":"+p)
		if err == nil {
			_ = ln.Close()
			return p
		}
	}
	return fallback
}

func main() {
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	if err := store.Init(context.TODO(), cfg.MongoURI, cfg.DBName); err != nil {
		log.Fatal(err)
	}

	handlers.SetConfig(handlers.ServerConfig{
		JWTSecret:          cfg.JWTSecret,
		ElevenLabsKey:      cfg.ElevenKeyBase,
		GeminiKey:          cfg.GeminiKeyBase,
		ChatGeminiKey:      cfg.GeminiKeyChat,
		ChatElevenKey:      cfg.ElevenKeyChat,
		NeighbourGeminiKey: cfg.GeminiKeyNeighbour,
		NeighbourElevenKey: cfg.ElevenKeyNeighbour,
		CloudName:          cfg.CloudName,
		CloudAPIKey:        cfg.CloudAPIKey,
		CloudAPISecret:     cfg.CloudAPISecret,
		FirestoreProjectID: cfg.FirestoreProjectID,
		GoogleCredentials:  cfg.GoogleCredentials,
	})

	if cfg.GoogleCredentials != "" {
		_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.GoogleCredentials)
	}
	if cfg.CloudName != "" {
		_ = os.Setenv("CLOUDINARY_CLOUD_NAME", cfg.CloudName)
	}
	if cfg.CloudAPIKey != "" {
		_ = os.Setenv("CLOUDINARY_API_KEY", cfg.CloudAPIKey)
	}
	if cfg.CloudAPISecret != "" {
		_ = os.Setenv("CLOUDINARY_API_SECRET", cfg.CloudAPISecret)
	}

	// Choose free ports to avoid conflicts
	httpPort := cfg.Port
	if httpPort == "" {
		httpPort = "8000"
	}
	httpPort = pickFreePort([]string{httpPort, "8080", "8088", "5050"}, httpPort)
	_ = os.Setenv("HTTP_PORT", httpPort)

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}
	metricsPort = pickFreePort([]string{metricsPort, "9191", "9292", "9393"}, metricsPort)
	_ = os.Setenv("METRICS_PORT", metricsPort)

	app := gofr.New()
	app.AddStaticFiles("/", "./public")

	app.POST("/signup", handlers.SignUp)
	app.POST("/login", handlers.Login)
	app.POST("/posts", handlers.CreatePost)
	app.GET("/feed", handlers.GetFeed)
	app.POST("/user/posts", handlers.GetUserPosts)

	app.POST("/api/signup", handlers.NeighbourSignUp)
	app.POST("/api/signin", handlers.NeighbourSignIn)
	app.POST("/api/upload", handlers.NeighbourUploadAudio)
	app.GET("/api/helper", handlers.NeighbourGetHelperRequests)
	app.POST("/api/assignRequest", handlers.NeighbourAssignRequest)
	app.POST("/api/eld-people/confirm", handlers.NeighbourConfirmRequest)
	app.POST("/api/reward/claim", handlers.NeighbourClaimReward)

	app.POST("/api/audio-chat", handlers.AudioChatHandler)

	log.Printf("Server running on port %s (metrics %s)", httpPort, metricsPort)
	app.Run()
}
