package extended

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Owner struct {
	UserID   int64 `json:"user_id"`
	VendorID int64 `json:"vendor_id"`
}

type OwnerModel struct {
	DB *sql.DB
}

func (u OwnerModel) GetVendorUserIds(vendorID int64) (*Owner, error) {
	if vendorID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT user_id, vendor_id
		FROM users_vendors
		WHERE vendor_id = $1
`

	var p Owner

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, vendorID).Scan(
		&p.UserID,
		&p.VendorID,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &p, nil

}

func (u OwnerModel) GetUserVendorIds(userID int64) (*Owner, error) {
	if userID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT user_id, vendor_id
		FROM users_vendors
		WHERE user_id = $1
`

	var owner Owner

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, userID).Scan(
		&owner.UserID,
		&owner.VendorID,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &owner, nil

}
