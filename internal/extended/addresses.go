package extended

import (
	"context"
	"database/sql"
	"github.com/Boostport/address"
	"github.com/hashicorp/go-multierror"
	"github.com/lib/pq"
	"time"
)

type Address struct {
	ID                 int64                  `json:"id"`
	CreatedAt          time.Time              `json:"created_at"`
	Country            string                 `json:"country"`
	Name               string                 `json:"name"`
	Organization       string                 `json:"organization"`
	StreetAddress      []string               `json:"street_address"`
	Locality           string                 `json:"locality"`
	AdministrativeArea string                 `json:"administrative_area"`
	PostCode           string                 `json:"post_code"`
	SortingCode        string                 `json:"sorting_code"`
	Data               map[string]interface{} `json:"-"`
	Lat                float64                `json:"lat"`
	Lng                float64                `json:"lng"`
	Coordinates        []float64              `json:"-"`
	DisplayName        string                 `json:"display_name"`
}

func ValidateAddress(a *Address) []any {
	addr, err := address.NewValid(
		address.WithCountry(a.Country), // Must be an ISO 3166-1 country code
		address.WithName(a.Name),
		address.WithOrganization(a.Organization),
		address.WithStreetAddress([]string{
			a.StreetAddress[0],
		}),
		address.WithLocality(a.Locality),
		address.WithAdministrativeArea(a.AdministrativeArea), // If the country has a pre-defined list of admin areas (like here), you must use the key and not the name
		address.WithPostCode(a.PostCode),
	)

	if err != nil {
		// If there was an error and you want to find out which validations failed,
		// type switch it as a *multierror.Error to access the list of errors
		if merr, ok := err.(*multierror.Error); ok {
			for _, subErr := range merr.Errors {
				if subErr == address.ErrInvalidCountryCode {
					// log.Fatalf(subErr)
				}
			}
		}
	}

	lang := "en"

	postalStringFormatter := address.PostalLabelFormatter{
		Output:            address.StringOutputter{},
		OriginCountryCode: "US", // We are sending from USA
	}
	psf := postalStringFormatter.Format(addr, lang)

	args := []any{addr, psf}

	return args

}

type AddressModel struct {
	DB *sql.DB
}

func (m AddressModel) Insert(address *Address) error {
	query := `
	INSERT INTO addresses (country,name,display_name,organization,street_address,locality,administrative_area,post_code,sorting_code,data,lat,lng,coordinates)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	RETURNING id, created_at, display_name;
	`
	args := []any{address.Country, address.Name, address.DisplayName, address.Organization, pq.Array(address.StreetAddress), address.Locality, address.AdministrativeArea, address.PostCode, address.SortingCode, address.Data, address.Lat, address.Lng, pq.Array(address.Coordinates)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&address.ID, &address.CreatedAt, &address.DisplayName)
}
