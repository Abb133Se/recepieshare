package model

type Recipe struct {
	ID          uint         `gorm:"primaryKey"`
	Title       string       `json:"title" binding:"required"`
	Text        string       `json:"text" binding:"required"`
	UserID      uint         `json:"user_id" binding:"required"`
	User        User         `gorm:"foreignKey:UserID" json:"-"`
	Ingridients []Ingridient `gorm:"foreignKey:RecipeID" json:"ingridients"`
	Comments    []Comment    `gorm:"foreignKey:RecipeID" json:"comments"`
}

type Ingridient struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `json:"name" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	RecipeID uint   `json:"recipe_id"`
}

type User struct {
	ID       uint      `gorm:"primaryKey"`
	Name     string    `json:"name"`
	LastName string    `json:"last_name"`
	Salt     string    `json:"salt"`
	Password string    `json:"password"`
	Email    string    `json:"email"`
	Comments []Comment `gorm:"foreignKey: UserID" json:"comments"`
	Recipes  []Recipe  `gorm:"foreignKey: UserID" json:"recipes"`
}

type Comment struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	UserID      uint   `json:"user_id"`
	RecipeID    uint   `json:"recipe_id"`
}
