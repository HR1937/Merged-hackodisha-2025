package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"finalapp/models"
	"finalapp/store"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gofr.dev/pkg/gofr"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("400: %v", err)
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("400: email/password required")
	}
	var existing models.User
	err := store.UsersCollection.FindOne(ctx.Request.Context(), bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		return nil, fmt.Errorf("409: user already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	user := models.User{Name: req.Name, Email: req.Email, Password: string(hash), CreatedAt: time.Now()}
	res, err := store.UsersCollection.InsertOne(ctx.Request.Context(), user)
	if err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	oid := res.InsertedID.(primitive.ObjectID)
	token, err := generateToken(oid)
	if err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	return map[string]interface{}{"user_id": oid.Hex(), "token": token}, nil
}

func Login(ctx *gofr.Context) (interface{}, error) {
	var req struct{ Email, Password string }
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("400: %v", err)
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	var user models.User
	err := store.UsersCollection.FindOne(ctx.Request.Context(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("404: invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("404: invalid credentials")
	}
	token, err := generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	return map[string]interface{}{"user_id": user.ID.Hex(), "token": token}, nil
}

func CreatePost(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Token     string   `json:"token"`
		MediaURL  string   `json:"media_url"`
		MediaType string   `json:"media_type"`
		Section   string   `json:"section"`
		Content   string   `json:"content"`
		Tags      []string `json:"tags"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("400: %v", err)
	}
	uidHex, err := parseToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("401: invalid token")
	}
	userOID, err := primitive.ObjectIDFromHex(uidHex)
	if err != nil {
		return nil, fmt.Errorf("400: invalid user ID")
	}
	var user models.User
	if err := store.UsersCollection.FindOne(ctx, bson.M{"_id": userOID}).Decode(&user); err != nil {
		return nil, fmt.Errorf("404: User not found")
	}
	textForHash := req.Content
	if req.MediaURL != "" && (req.MediaType == "audio" || req.MediaType == "video") {
		if transcript, err := TranscribeElevenLabs(req.MediaURL); err == nil {
			textForHash = transcript
		}
	}
	var hashtags []string
	if textForHash != "" {
		if tags, err := GenerateHashtags(context.Background(), textForHash); err == nil {
			hashtags = tags
		}
	}
	post := models.Post{UserID: uidHex, UserName: user.Name, MediaURL: req.MediaURL, MediaType: req.MediaType, Content: req.Content, Tags: hashtags, Section: req.Section, CreatedAt: time.Now()}
	res, err := store.PostsCollection.InsertOne(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	post.ID = res.InsertedID.(primitive.ObjectID)
	return post, nil
}

func GetFeed(ctx *gofr.Context) (interface{}, error) {
	cur, err := store.PostsCollection.Find(ctx.Request.Context(), bson.M{}, nil)
	if err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	defer cur.Close(ctx.Request.Context())
	var posts []models.Post
	if err := cur.All(ctx.Request.Context(), &posts); err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	return posts, nil
}

func GetUserPosts(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Token string `json:"token"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("400: %v", err)
	}
	uidHex, err := parseToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("401: invalid token")
	}
	section := ctx.Param("section")
	filter := bson.M{"userid": uidHex}
	if section != "" {
		filter = bson.M{"userid": uidHex, "section": section}
	}
	cur, err := store.PostsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	defer cur.Close(ctx)
	var posts []models.Post
	if err := cur.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("500: %v", err)
	}
	return posts, nil
}

func generateToken(userID primitive.ObjectID) (string, error) {
	claims := jwt.MapClaims{"user_id": userID.Hex(), "exp": time.Now().Add(24 * time.Hour).Unix()}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(cfg.JWTSecret))
}

func parseToken(tok string) (string, error) {
	if tok == "" {
		return "", fmt.Errorf("token missing")
	}
	parsed, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) { return []byte(cfg.JWTSecret), nil })
	if err != nil {
		return "", err
	}
	if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		if uid, ok := claims["user_id"].(string); ok {
			return uid, nil
		}
	}
	return "", fmt.Errorf("invalid token claims")
}
