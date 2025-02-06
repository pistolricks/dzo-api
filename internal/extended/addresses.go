package extended

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Boostport/address"
	"github.com/lib/pq"
	"log"
	"time"
)

type Address struct {
	ID                 int64                  `json:"id"`
	CreatedAt          time.Time              `json:"created_at"`
	Country            string                 `json:"country"`
	Name               string                 `json:"name,omitempty"`
	Organization       string                 `json:"organization,omitempty"`
	StreetAddress      []string               `json:"street_address"`
	Locality           string                 `json:"locality"`
	AdministrativeArea string                 `json:"administrative_area"`
	PostCode           string                 `json:"post_code"`
	SortingCode        string                 `json:"sorting_code"`
	Data               map[string]interface{} `json:"-"`
	Lat                float64                `json:"lat,omitempty"`
	Lng                float64                `json:"lng,omitempty"`
}

func ValidateAddress(a *Address) []any {
	addr, err := address.NewValid(
		address.WithCountry(a.Country), // Must be an ISO 3166-1 country code
		address.WithName(a.Name),
		address.WithOrganization(a.Organization),
		/*
			address.WithStreetAddress([]string{
				"525 Collins Street",
			}),
		*/
		address.WithStreetAddress(a.StreetAddress),
		address.WithLocality(a.Locality),
		address.WithAdministrativeArea(a.AdministrativeArea), // If the country has a pre-defined list of admin areas (like here), you must use the key and not the name
		address.WithPostCode(a.PostCode),
	)

	if err != nil {
		// If there was an error and you want to find out which validations failed, use errors.Is()
		if errors.Is(err, address.ErrInvalidCountryCode) {
			log.Fatalf("Invalid country code")
		}
	}

	return []any{addr}
}

type AddressModel struct {
	DB *sql.DB
}

func (m AddressModel) Insert(address *Address) error {
	query := `
	INSERT INTO addresses (country,name,organization,street_address,locality,administrative_area,post_code,sorting_code,data,lat,lng)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id, created_at, name;
	`
	args := []any{address.Country, address.Name, address.Organization, pq.Array(address.StreetAddress), address.Locality, address.AdministrativeArea, address.PostCode, address.SortingCode, address.Data, address.Lat, address.Lng}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&address.ID, &address.CreatedAt, &address.Name)
}
