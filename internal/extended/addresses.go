package extended

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Boostport/address"
	"github.com/indrasaputra/hashids"
	"github.com/lib/pq"
	"github.com/muesli/gominatim"
	"github.com/turistikrota/osm"
	"time"
)

type Address struct {
	ID                 hashids.ID             `json:"id"`
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

func ValidateAddress(a *Address) (address.Address, error) {
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

	return addr, err
}

func SearchOsm(s string) ([]osm.SearchResult, error) {
	ctx := context.Background()
	results, err := osm.Search(ctx, s)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}
	for _, result := range results {
		fmt.Println("Search Result:", result)
	}

	return results, err
}

func GetDetailsWithPlaceId(id int) (*osm.DetailsResult, error) {
	ctx := context.Background()
	result, err := osm.DetailsWithPlaceID(ctx, id) // PlaceID for the Empire State Building
	if err != nil {
		fmt.Println("Error:", err)
	}
	return result, err
}

func GetDetailsWithCoordinates(lat float64, long float64) (*osm.ReverseResult, error) {
	ctx := context.Background()
	result, err := osm.Reverse(ctx, lat, long) // Latitude and Longitude for the Empire State Building
	if err != nil {
		fmt.Println("Error:", err)
	}
	return result, err
}

func GetAddressOSM(a address.Address) ([]gominatim.SearchResult, error) {
	gominatim.SetServer("https://nominatim.openstreetmap.org/")

	//Get by a Querystring

	//Get by City
	qry := gominatim.SearchQuery{
		Q:          a.StreetAddress[0],
		City:       a.Locality,
		State:      a.AdministrativeArea,
		Postalcode: a.PostCode,
	}
	resp, qer := qry.Get() // Returns []gominatim.SearchResult

	return resp, qer
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
