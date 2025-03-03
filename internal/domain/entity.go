package domain

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	ID            uuid.UUID  `gorm:"primaryKey;type:uuid"`
	Text          string     `gorm:"type:text;not null"`
	AllowComments bool       `gorm:"type:boolean;not null;default:true"`
	Comments      []*Comment `gorm:"foreignKey:PostID"`
	CreatedAt     *time.Time `gorm:"type:timestamp with time zone;not null"`
}

type Comment struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid"`
	Text      string     `gorm:"type:text;not null;check:length(text) <= 2000"`
	PostID    uuid.UUID  `gorm:"type:uuid;not null"`
	ParentID  *uuid.UUID `gorm:"type:uuid"`
	CreatedAt *time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	Children  []*Comment `gorm:"foreignKey:ParentID"`
}

func (Post) TableName() string {
	return "posts"
}

func (Comment) TableName() string {
	return "comments"
}
