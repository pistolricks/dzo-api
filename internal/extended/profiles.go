package extended

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type ProfileDB interface {
	CreateProfile(profileName string, name string, hash string, data []byte) error
	DeleteProfile(hash string) error
	GetProfileKey(hash string, keys ...string) ([]byte, error)
	CreateProfileKey(hash string, data []byte, keys ...string) error
	UpdateProfileKey(hash string, data []byte, keys ...string) error
	DeleteProfileKey(hash string, keys ...string) error
}

type Profile struct {
	ID        string    `json:"id"`
	Profile   string    `json:"profile"`
	Name      string    `json:"name"`
	Hash      string    `json:"hash"`
	Data      Attrs     `json:"data"`
	CreatedAt time.Time `json:"-"`
}

type Attrs map[string]interface{}

func (a Attrs) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Attrs) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type ProfileModel struct {
	DB *sql.DB
}

func (m ProfileModel) Insert(profile *Profile) error {

	item := new(Profile)

	_, err := m.DB.Exec("INSERT INTO trees (profile,name, hash, data) VALUES($1,$2,$3)", profile.Profile, profile.Name, profile.Hash, profile.Data)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	item = new(Profile)

	return m.DB.QueryRowContext(ctx, "SELECT id, profile, name, hash, data, created_at FROM trees ORDER BY profile DESC LIMIT 1").Scan(&item.ID, &item.Profile, &item.Name, &item.Hash, &item.Data, &item.CreatedAt)
}

func (m ProfileModel) GetProjectKey(Name string, keys ...string) ([]byte, error) {
	byt := []byte{}

	keysFormat := strings.Join(keys, ",")
	err := m.DB.QueryRow(
		fmt.Sprintf("SELECT data#>'{%s}' as data FROM trees WHERE profile=$1 ORDER BY id DESC LIMIT 1", keysFormat),
		Name).Scan(&byt)

	return byt, err
}

func (m ProfileModel) CreateProjectKey(name string, data []byte, keys ...string) error {
	keysFormat := strings.Join(keys, ",")
	_, err := m.DB.Exec(fmt.Sprintf("UPDATE trees set data=jsonb_insert(data, '{%s}', $1) WHERE name=$2", keysFormat), data, name)
	return err
}

func (m ProfileModel) UpdateProjectKey(name string, data []byte, keys ...string) error {
	keysFormat := strings.Join(keys, ",")
	_, err := m.DB.Exec(fmt.Sprintf("UPDATE trees set data=jsonb_set(data, '{%s}', $1) WHERE name=$2", keysFormat), data, name)
	return err
}

// DeleteProjectKey permanently removes the data at the key path.
func (m ProfileModel) DeleteProjectKey(name string, keys ...string) error {
	keysFormat := strings.Join(keys, ",")
	_, err := m.DB.Exec(fmt.Sprintf("UPDATE trees SET data=data #- '{%s}' where name=$1", keysFormat), name)
	return err
}

func (m ProfileModel) Delete(hashId string) error {
	_, err := m.DB.Exec("DELETE FROM trees where hash=$1", hashId)
	return err
}

func (m ProfileModel) Get(hash string) (*Profile, error) {

	query := `
		SELECT id, profile, name, hash, data, created_at
		FROM trees
		WHERE hash = $1`

	var profile Profile

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, hash).Scan(
		&profile.ID,
		&profile.CreatedAt,
		&profile.ID,
		&profile.Profile,
		&profile.Name,
		&profile.Hash,
		&profile.Data,
		&profile.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &profile, nil
}

func (m ProfileModel) GetAll(profileName string, name string, hashId string, dataArr Attrs, filters Filters) ([]*Profile, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, profile, name, hash, data, created_at
FROM trees
WHERE (to_tsvector('simple', profile) @@ plainto_tsquery('simple', $1) OR $1 = '')
  AND (to_tsvector('simple', name) @@ plainto_tsquery('simple', $2) OR $2 = '')
  AND (data::jsonb @> $3::jsonb OR $3 = '{}')
ORDER BY profile ASC, created_at ASC
LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{profileName, name, dataArr, filters.limit(), filters.offset()}
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
	profiles := []*Profile{}

	for rows.Next() {
		var profile Profile

		err := rows.Scan(
			&totalRecords,
			&profile.ID,
			&profile.Profile,
			&profile.Name,
			&profile.Hash,
			&profile.Data,
			&profile.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		profiles = append(profiles, &profile)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return profiles, metadata, nil
}
