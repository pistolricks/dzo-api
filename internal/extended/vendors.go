package extended

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/pistolricks/validation"
	"time"
)

type Vendor struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateVendor(v *validation.Validator, vendor *Vendor) {
	v.Check(vendor.Title != "", "title", "is required")

	v.Check(vendor.Year >= 1888, "year", "must be greater than or equal to 1888")
	v.Check(vendor.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(vendor.Runtime != 0, "runtime", "is required")
	v.Check(vendor.Runtime > 0, "runtime", "must be a positive number")

	v.Check(vendor.Genres != nil, "genres", "must be provided")
	v.Check(len(vendor.Genres) >= 1, "genres", "must contain at least one")
	v.Check(len(vendor.Genres) <= 5, "genres", "must contain no more than five")
	v.Check(validation.Unique(vendor.Genres), "genres", "must not contain duplicate values")

}

type VendorModel struct {
	DB *sql.DB
}

func (m VendorModel) Insert(vendor *Vendor) error {
	query := `
	INSERT INTO vendors (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version;
	`
	args := []any{vendor.Title, vendor.Year, vendor.Runtime, pq.Array(vendor.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&vendor.ID, &vendor.CreatedAt, &vendor.Version)
}

func (m VendorModel) Get(id int64) (*Vendor, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM vendors
		WHERE id = $1`

	var vendor Vendor

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&vendor.ID,
		&vendor.CreatedAt,
		&vendor.Title,
		&vendor.Year,
		&vendor.Runtime,
		pq.Array(&vendor.Genres),
		&vendor.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &vendor, nil
}

func (m VendorModel) Update(vendor *Vendor) error {
	query := `
		UPDATE vendors
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{
		vendor.Title,
		vendor.Year,
		vendor.Runtime,
		pq.Array(vendor.Genres),
		vendor.ID,
		vendor.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&vendor.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}

	}
	return nil
}

func (m VendorModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM vendors
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m VendorModel) GetAll(title string, genres []string, filters Filters) ([]*Vendor, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
	FROM vendors
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}
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
	vendors := []*Vendor{}

	for rows.Next() {
		var vendor Vendor

		err := rows.Scan(
			&totalRecords,
			&vendor.ID,
			&vendor.CreatedAt,
			&vendor.Title,
			&vendor.Year,
			&vendor.Runtime,
			pq.Array(&vendor.Genres),
			&vendor.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		vendors = append(vendors, &vendor)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return vendors, metadata, nil
}
