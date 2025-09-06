package model

import "time"

type Recipe struct {
	ID          uint         `gorm:"primaryKey"`
	Title       string       `json:"title" binding:"required"`
	Text        string       `json:"text" binding:"required"`
	UserID      uint         `json:"user_id" binding:"required"`
	User        User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Ingredients []Ingredient `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"ingredients"`
	Comments    []Comment    `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"comments"`
	Favorites   []Favorite   `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"favorites"`
	Ratings     []Rating     `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"ratings"`
	Tags        []Tag        `gorm:"many2many:recipe_tags;constraint:OnDelete:CASCADE" json:"tags"`
	Categories  []Category   `gorm:"many2many:recipe_categories;constraint:OnDelete:CASCADE" json:"categories"`
	Steps       []Step       `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"steps"`
	Calories    float64      `json:"calories"`
	Protein     float64      `json:"protein"`
	Fat         float64      `json:"fat"`
	Carbs       float64      `json:"carbs"`
	Fiber       float64      `json:"fiber"`
	Sugar       float64      `json:"sugar"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
}

type Ingredient struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required"`
	Amount    string    `json:"amount" binding:"required"`
	RecipeID  uint      `json:"recipe_id" gorm:"index"`
	Calories  float64   `json:"calories"`
	Protein   float64   `json:"protein"`
	Fat       float64   `json:"fat"`
	Carbs     float64   `json:"carbs"`
	Fiber     float64   `json:"fiber"`
	Sugar     float64   `json:"sugar"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Step struct {
	ID        uint      `gorm:"primaryKey"`
	RecipeID  uint      `json:"recipe_id" gorm:"index"`
	Order     int       `json:"order" binding:"required"`
	Text      string    `json:"text" binding:"required"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type User struct {
	ID                     uint       `gorm:"primaryKey"`
	Name                   string     `json:"name"`
	LastName               string     `json:"last_name"`
	Salt                   string     `json:"salt"`
	Password               string     `json:"password"`
	Email                  string     `json:"email"`
	Role                   string     `gorm:"deafault:user" json:"role"`
	PasswordResetToken     string     `json:"-" gorm:"size:255"`
	PasswordResetExpiresAt *time.Time `json:"-"`
	Comments               []Comment  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments"`
	Recipes                []Recipe   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"recipes"`
	Favorites              []Favorite `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"favorites"`
	Ratings                []Rating   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"ratings"`
	CreatedAt              time.Time  `gorm:"autoCreateTime"`
	UpdatedAt              time.Time  `gorm:"autoUpdateTime"`
}

type Comment struct {
	ID          uint      `gorm:"primaryKey"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Likes       int       `gorm:"default:0" json:"likes"`
	UserID      uint      `json:"user_id" gorm:"index"`
	RecipeID    uint      `json:"recipe_id" gorm:"index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type Favorite struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `json:"user_id" binding:"required" gorm:"index"`
	RecipeID  uint      `json:"recipe_id" binding:"required" gorm:"index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Rating struct {
	ID        uint      `gorm:"primaryKey"`
	RecipeID  uint      `json:"recipe_id" gorm:"index"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Score     uint      `gorm:"check:score >= 1 AND score <= 5" json:"score" binding:"required"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Tag struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required" gorm:"unique;not null"`
	Recipes   []Recipe  `gorm:"many2many:recipe_tags;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Category struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required" gorm:"unique;not null"`
	Recipes   []Recipe  `gorm:"many2many:recipe_categories;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Image struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EntityType string    `gorm:"index;size:50" json:"entity_type"`
	EntityID   uint      `gorm:"index" json:"entity_id"`
	Path       string    `json:"path"`
	Format     string    `json:"format"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
}
