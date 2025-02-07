package extended

import (
	"database/sql"
	"encoding/hex"
	"github.com/devedge/imagehash"
	"github.com/pistolricks/validation"
	"time"
)

type Content struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name,omitempty"`
	Hash      string    `json:"hash,omitempty"`
	Src       string    `json:"src"`
	Type      string    `json:"type,omitempty"`
	Size      int32     `json:"size,omitempty"`
	UserID    string    `json:"user_id"`
}

func ValidateContent(v *validation.Validator, content *Content) {
	v.Check(content.Name != "", "name", "is required")
	v.Check(content.Size > 0, "size", "This content doesn't have any data to it")
}

func HashImage(image string) string {
	src, _ := imagehash.OpenImg(image)
	hashLen := 8
	hash, _ := imagehash.Dhash(src, hashLen)

	return hex.EncodeToString(hash)
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
