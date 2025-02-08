package extended

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/devedge/imagehash"
	"github.com/indrasaputra/hashids"
	"github.com/pistolricks/validation"
	"os"
	"strings"
	"time"
)

var (
	ErrDuplicateHash = errors.New("duplicate hash")
)

type Content struct {
	ID        hashids.ID `json:"id"`
	CreatedAt time.Time  `json:"-"`
	Name      string     `json:"name,omitempty"`
	Original  string     `json:"original,omitempty"`
	Hash      string     `json:"hash,omitempty"`
	Src       string     `json:"src"`
	Type      string     `json:"type,omitempty"`
	Size      int64      `json:"size,omitempty"`
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

func (m ContentModel) Insert(content *Content) error {

	content.Src = strings.ReplaceAll(content.Src, "ui/static", "static")

	query := `
	INSERT INTO contents (name,original,hash,src,type,size,user_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at, name;
	`
	args := []any{content.Name, content.Original, content.Hash, content.Src, content.Type, content.Size, content.UserID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&content.ID, &content.CreatedAt, &content.Name)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "contents_hash_key"`:
			return ErrDuplicateHash
		default:
			return err
		}
	}
	return err
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

func (m ContentModel) GetAll(name string, original string, src string, mimeType string, size int64, userId string, filters Filters) ([]*Content, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, name, original, src, type, size, user_id
	FROM contents
	WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	ORDER BY %s %s, id ASC
	LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{name, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	totalRecords := 0
	contents := []*Content{}

	for rows.Next() {

		var content Content
		err := rows.Scan(
			&totalRecords,
			&content.ID,
			&content.Name,
			&content.Original,
			&content.Src,
			&content.Type,
			&content.Size,
			&content.UserID,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		contents = append(contents, &content)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return contents, metadata, nil
}

func OpenOutputFile(name string) (fp *os.File) {
	fp, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := fp.Close(); err != nil {
			panic(err)
		}
	}()

	return fp
}
