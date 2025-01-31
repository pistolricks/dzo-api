package data

import (
	"database/sql"
	"github.com/pistolricks/validation"
	"time"
)

type Content struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name,omitempty"`
	Src       string    `json:"src"`
	Type      string    `json:"type,omitempty"`
	Size      int32     `json:"size,omitempty"`
	Width     float32   `json:"width"`
	Height    float32   `json:"height"`
	SortOrder int16     `json:"sort_order"`
	UserID    string    `json:"user_id"`
}

func ValidateContent(v *validation.Validator, content *Content) {
	v.Check(content.Name != "", "name", "is required")
	v.Check(content.Size > 0, "size", "This content doesn't have any data to it")
	v.Check(content.SortOrder > 0, "sort_order", "order must be greater than zero")
}

type ContentModel struct {
	DB *sql.DB
}

func (m ContentModel) EncodeWebP(content *Content) error {

	return nil
}

func (m ContentModel) DecodeWebP(content *Content) error {

	return nil
}
