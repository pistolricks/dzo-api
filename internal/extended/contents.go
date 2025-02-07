package extended

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/devedge/imagehash"
	"github.com/indrasaputra/hashids"
	"github.com/pistolricks/validation"
	"time"
)

type Content struct {
	ID        hashids.ID `json:"id"`
	CreatedAt time.Time  `json:"-"`
	Name      string     `json:"name,omitempty"`
	Original  string     `json:"original,omitempty"`
	Hash      string     `json:"hash,omitempty"`
	Src       string     `json:"src"`
	Type      string     `json:"type,omitempty"`
	Size      int32      `json:"size,omitempty"`
	UserID    string     `json:"user_id"`
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

func (m ContentModel) GetByHash(hash string) (*Content, error) {
	query := `
	SELECT id,created_at,name,original,hash,src,type,size,user_id
FROM contents
WHERE hash = $1`
	var content Content

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, hash).Scan(
		&content.ID,
		&content.CreatedAt,
		&content.Name,
		&content.Original,
		&content.Hash,
		&content.Src,
		&content.Type,
		&content.Size,
		&content.UserID,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &content, nil
}

func (m ContentModel) EncodeWebP(content *Content) error {

	return nil
}

func (m ContentModel) DecodeWebP(content *Content) error {

	return nil
}
