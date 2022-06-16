package DB


type User struct {
	Name     string `gorm:"name" json:"name"`
	Password string `gorm:"password" json:"password"`
	Role     string `gorm:"role" json:"role"`
}

type Book struct {
	ID			   int    `gorm:"id" json:"id"`
	Name           string `gorm:"name" json:"name"`
	Author         string `gorm:"author" json:"author"`
	Floor          int    `gorm:"floor" json:"floor"`
	Block          string `gorm:"block" json:"block"`
	Bookshelf      int    `gorm:"bookshelf" json:"bookshelf"`
	BookshelfLevel int    `gorm:"bookshelf_level" json:"bookshelf_level"`
	Exist          bool   `gorm:"exist" json:"exist"`
}