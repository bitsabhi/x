package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Name     string
	Email    string
	Password string // Securely store the user's hashed password
}

type News struct {
	gorm.Model
	Title    string
	Content  string
	Category string
	Source   string
	URL      string
}

type UserPreference struct {
	gorm.Model
	UserID    uint
	Category  string
	Frequency int
}

type UserInteraction struct {
	gorm.Model
	UserID   uint
	NewsID   uint
	Action   string // e.g., "click", "like"
	Duration int    // Time spent in seconds
}
