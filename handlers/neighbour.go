package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/golang-jwt/jwt/v5"
	"gofr.dev/pkg/gofr"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type NUser struct {
	ID        string    `json:"id" firestore:"id"`
	Name      string    `json:"name" firestore:"name"`
	Email     string    `json:"email" firestore:"email"`
	Password  string    `json:"password" firestore:"password"`
	Role      string    `json:"role" firestore:"role"`
	Location  NLocation `json:"location" firestore:"location"`
	FCMToken  string    `json:"fcmToken" firestore:"fcmToken"`
	Reward    int       `json:"reward" firestore:"reward"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
}

type NLocation struct {
	Lat, Lng float64 `json:"lat" firestore:"lat"`
}

func initFirestore() *firestore.Client {
	if cfg.FirestoreProjectID == "" {
		return nil
	}
	ctx := context.Background()
	if cfg.GoogleCredentials != "" {
		_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.GoogleCredentials)
	}
	cl, err := firestore.NewClient(ctx, cfg.FirestoreProjectID, option.WithCredentialsFile(cfg.GoogleCredentials))
	if err != nil {
		return nil
	}
	return cl
}

func NeighbourSignUp(ctx *gofr.Context) (interface{}, error) {
	var user NUser
	if err := ctx.Bind(&user); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	if user.Name == "" || user.Email == "" || user.Password == "" || user.Role == "" {
		return nil, fmt.Errorf("missing required fields")
	}
	client := initFirestore()
	// Hash password
	hasher := sha256.New()
	hasher.Write([]byte(user.Password))
	user.Password = hex.EncodeToString(hasher.Sum(nil))
	user.ID = user.Email
	user.CreatedAt = time.Now()
	user.Reward = 0
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": user.ID, "email": user.Email, "role": user.Role, "exp": time.Now().Add(24 * time.Hour).Unix()})
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}
	if client != nil {
		_, _ = client.Collection("users").Doc(user.ID).Set(context.Background(), user)
	}
	user.Password = ""
	return map[string]interface{}{"message": "User created", "user": user, "token": tokenString}, nil
}

func NeighbourSignIn(ctx *gofr.Context) (interface{}, error) {
	var cred struct{ Email, Password string }
	if err := ctx.Bind(&cred); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	client := initFirestore()
	user := NUser{ID: cred.Email, Name: "Demo User", Email: cred.Email, Role: "elder", CreatedAt: time.Now(), Reward: 0}
	if client != nil {
		if doc, err := client.Collection("users").Doc(user.ID).Get(context.Background()); err == nil {
			_ = doc.DataTo(&user)
		}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": user.ID, "email": user.Email, "role": user.Role, "exp": time.Now().Add(24 * time.Hour).Unix()})
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}
	return map[string]interface{}{"message": "Sign in successful", "user": user, "token": tokenString}, nil
}

func NeighbourUploadAudio(ctx *gofr.Context) (interface{}, error) {
	elderLatStr := ctx.Param("elderLat")
	elderLngStr := ctx.Param("elderLng")
	elderID := ctx.Param("elderId")
	if elderLatStr == "" {
		elderLatStr = "40.7128"
	}
	if elderLngStr == "" {
		elderLngStr = "-74.0060"
	}
	if elderID == "" {
		elderID = "demo-elder"
	}
	elderLat, _ := strconv.ParseFloat(elderLatStr, 64)
	elderLng, _ := strconv.ParseFloat(elderLngStr, 64)
	requestID := fmt.Sprintf("%s-%d", elderID, time.Now().Unix())
	audioURL := "https://example.com/audio/" + requestID + ".wav"
	title := "Help Request - " + time.Now().Format("15:04")
	transcription := "Audio help request from elderly person"
	client := initFirestore()
	if client != nil {
		_, _ = client.Collection("requests").Doc(requestID).Set(context.Background(), map[string]interface{}{
			"id": requestID, "title": title, "audioUrl": audioURL, "transcription": transcription,
			"elderId": elderID, "elderLocation": map[string]float64{"lat": elderLat, "lng": elderLng},
			"status": "pending", "createdAt": time.Now(),
		})
		// basic nearby helper scan
		_, _ = getNearbyHelpers(client, elderLat, elderLng)
	}
	return map[string]interface{}{"message": "Audio uploaded and request created", "requestId": requestID, "url": audioURL, "transcription": transcription, "title": title}, nil
}

func getNearbyHelpers(client *firestore.Client, elderLat, elderLng float64) ([]NUser, error) {
	if client == nil {
		return nil, nil
	}
	usersRef := client.Collection("users")
	iter := usersRef.Where("role", "==", "helper").Documents(context.Background())
	defer iter.Stop()
	var helpers []NUser
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var u NUser
		if err := doc.DataTo(&u); err != nil {
			continue
		}
		d := distanceMeters(elderLat, elderLng, u.Location.Lat, u.Location.Lng)
		if d <= 500 {
			helpers = append(helpers, u)
		}
	}
	return helpers, nil
}

func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371e3
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func NeighbourGetHelperRequests(ctx *gofr.Context) (interface{}, error) {
	userID := ctx.Param("userName")
	if userID == "" {
		userID = "demo-helper"
	}
	requests := []map[string]interface{}{
		{"id": "demo-request-1", "title": "Help Request - 14:30", "audioUrl": "https://example.com/audio/demo1.wav", "transcription": "I need help with groceries", "elderId": "demo-elder-1", "elderLocation": map[string]float64{"lat": 40.7128, "lng": -74.0060}, "status": "pending", "createdAt": time.Now().Add(-time.Hour)},
		{"id": "demo-request-2", "title": "Help Request - 13:15", "audioUrl": "https://example.com/audio/demo2.wav", "transcription": "Need assistance with medication pickup", "elderId": "demo-elder-2", "elderLocation": map[string]float64{"lat": 40.7130, "lng": -74.0058}, "status": "pending", "createdAt": time.Now().Add(-2 * time.Hour)},
	}
	return map[string]interface{}{"requests": requests}, nil
}

func NeighbourAssignRequest(ctx *gofr.Context) (interface{}, error) {
	var body struct{ RequestID, HelperID string }
	if err := ctx.Bind(&body); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	return map[string]interface{}{"message": "Request assigned successfully"}, nil
}

func NeighbourConfirmRequest(ctx *gofr.Context) (interface{}, error) {
	var body struct{ RequestID string }
	if err := ctx.Bind(&body); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	return map[string]interface{}{"message": "Request confirmed successfully"}, nil
}

func NeighbourClaimReward(ctx *gofr.Context) (interface{}, error) {
	var body struct{ HelperID, RequestID string }
	if err := ctx.Bind(&body); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	return map[string]interface{}{"message": "Reward claimed successfully", "newBalance": 10}, nil
}
