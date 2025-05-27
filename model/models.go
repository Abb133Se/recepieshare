package model

import "time"

type Recipe struct {
	ID          uint         `gorm:"primaryKey"`
	Title       string       `json:"title" binding:"required"`
	Text        string       `json:"text" binding:"required"`
	UserID      uint         `json:"user_id" binding:"required"`
	User        User         `gorm:"foreignKey:UserID" json:"-"`
	Ingridients []Ingridient `gorm:"foreignKey:RecipeID" json:"ingridients"`
	Comments    []Comment    `gorm:"foreignKey:RecipeID" json:"comments"`
	Favorited   []Favorite   `gorm:"foreignKey: RecipeID" json:"favorites"`
	Ratings     []Rating     `gorm:"foreignKey:RecipeID" json:"ratings"`
}

type Ingridient struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `json:"name" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	RecipeID uint   `json:"recipe_id"`
}

type User struct {
	ID                     uint       `gorm:"primaryKey"`
	Name                   string     `json:"name"`
	LastName               string     `json:"last_name"`
	Salt                   string     `json:"salt"`
	Password               string     `json:"password"`
	Email                  string     `json:"email"`
	PasswordResetToken     string     `json:"-" gorm:"size:255"`
	PasswordResetExpiresAt *time.Time `json:"-"`
	Comments               []Comment  `gorm:"foreignKey: UserID" json:"comments"`
	Recipes                []Recipe   `gorm:"foreignKey: UserID" json:"recipes"`
	Favorites              []Favorite `gorm:"foreignKey: UserID" json:"favorites"`
	Ratings                []Rating   `gorm:"foreignKey:UserID" json:"ratings"`
}

type Comment struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Likes       int    `gorm:"default:0" json:"likes"`
	UserID      uint   `json:"user_id"`
	RecipeID    uint   `json:"recipe_id"`
}

type Favorite struct {
	ID       uint `gorm:"primaryKey"`
	UserID   uint `json:"user_id" binding:"required"`
	RecipeID uint `json:"recipe_id" binding:"required"`
}

type Rating struct {
	ID       uint `gorm:"primaryKey"`
	RecipeID uint `json:"recipe_id"`
	UserID   uint `json:"user_id"`
	Score    uint `gorm:"check:score >= 1 AND score <= 5" json:"score" binding:"required"`
}
