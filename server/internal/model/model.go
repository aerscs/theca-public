package model

type User struct {
	Email               string `json:"email" gorm:"size:255;unique;not null;index:idx_users_email"`
	Username            string `json:"username" gorm:"size:255;unique;not null;index:idx_users_username"`
	PassHash            string `json:"-" gorm:"size:255;not null"`
	ID                  uint   `json:"id" gorm:"primary_key;unique;not null"`
	RefreshTokenVersion uint   `json:"-" gorm:"default:0"`
	AmountOfBookmarks   uint   `json:"amount_of_bookmarks" gorm:"default:0"`
	IsVerified          bool   `json:"-" gorm:"default:false;index:idx_users_is_verified"`
	IsPremium           bool   `json:"is_premium" gorm:"default:false"`
}
