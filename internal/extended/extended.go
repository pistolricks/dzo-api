package extended

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Extended struct {
	Vendors   VendorModel
	Contents  ContentModel
	Addresses AddressModel
	Profiles  ProfileModel
	Owners    OwnerModel
	Users     UserModel
}

func NewExtended(db *sql.DB) Extended {
	return Extended{
		Vendors:   VendorModel{DB: db},
		Contents:  ContentModel{DB: db},
		Addresses: AddressModel{DB: db},
		Profiles:  ProfileModel{DB: db},
		Owners:    OwnerModel{DB: db},
		Users:     UserModel{DB: db},
	}
}
