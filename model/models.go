package model

import "time"

type Recipe struct {
	ID          uint         `gorm:"primaryKey"`
	Title       string       `json:"title" binding:"required"`
	Text        string       `json:"text" binding:"required"`
	UserID      uint         `json:"user_id" binding:"required"`
	User        User         `gorm:"foreignKey:UserID" json:"-"`
	Ingredients []Ingredient `gorm:"foreignKey:RecipeID" json:"ingredients"`
	Comments    []Comment    `gorm:"foreignKey:RecipeID" json:"comments"`
	Favorited   []Favorite   `gorm:"foreignKey: RecipeID" json:"favorites"`
	Ratings     []Rating     `gorm:"foreignKey:RecipeID" json:"ratings"`
	Tags        []Tag        `gorm:"many2many:recipe_tags" json:"tags"`
	Categories  []Category   `gorm:"many2many:recipe_categories" json:"categories"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Ingredient struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `json:"name" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	RecipeID  uint   `json:"recipe_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
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
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type Comment struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Likes       int    `gorm:"default:0" json:"likes"`
	UserID      uint   `json:"user_id"`
	RecipeID    uint   `json:"recipe_id"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Favorite struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `json:"user_id" binding:"required"`
	RecipeID  uint `json:"recipe_id" binding:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Rating struct {
	ID        uint `gorm:"primaryKey"`
	RecipeID  uint `json:"recipe_id"`
	UserID    uint `json:"user_id"`
	Score     uint `gorm:"check:score >= 1 AND score <= 5" json:"score" binding:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tag struct {
	ID        uint     `gorm:"primaryKey"`
	Name      string   `json:"name" binding:"required" gorm:"unique;not null"`
	Recipes   []Recipe `gorm:"many2many:recipe_tags" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Category struct {
	ID        uint     `gorm:"primaryKey"`
	Name      string   `json:"name" binding:"required" gorm:"unique;not null"`
	Recipes   []Recipe `gorm:"many2many:recipe_categories" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
