package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email          string             `bson:"email" json:"email"`
	Password       string             `bson:"password_hash" json:"-"` // Changed to match DB field
	Name           string             `bson:"name" json:"name"`
	Bio            string             `bson:"bio" json:"bio"`
	Location       string             `bson:"location" json:"location"`
	ProfilePicture string             `bson:"profile_picture" json:"profile_picture"`
	LikedPosts     []string           `bson:"liked_posts" json:"liked_posts"`
	PreferredTags  []string           `bson:"preferred_tags" json:"preferred_tags"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"userid" json:"user_id"`     // Changed to match handler usage
	UserName  string             `bson:"username" json:"user_name"` // Added UserName field
	Content   string             `bson:"content" json:"content"`    // Added Content field
	MediaURL  string             `bson:"media_url" json:"media_url"`
	MediaType string             `bson:"media_type" json:"media_type"`
	Section   string             `bson:"section" json:"section"`
	Tags      []string           `bson:"tags" json:"tags"`
	Likes     int                `bson:"likes" json:"likes"`
	LikedBy   []string           `bson:"liked_by" json:"liked_by"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type ProfileUpdateRequest struct {
	Name           string `json:"name"`
	Bio            string `json:"bio"`
	Location       string `json:"location"`
	ProfilePicture string `json:"profile_picture"`
}
