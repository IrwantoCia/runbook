package folder

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound  = errors.New("folder not found")
	ErrNotEmpty  = errors.New("folder still contains runbooks or subfolders")
	ErrCycle     = errors.New("cannot move a folder into itself or its subfolders")
	ErrDuplicate = errors.New("a folder with this name already exists here")
	ErrInvalid   = errors.New("invalid folder name or parent")
)

type Folder struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

type Repository struct{ DB *sql.DB }

func validName(name string) bool {
	name = strings.TrimSpace(name)
	return name != "" && len([]rune(name)) <= 80 && !strings.ContainsAny(name, "/\\\x00")
}

func (r Repository) List() ([]Folder, error) {
	rows, err := r.DB.Query(`SELECT id,name,parent_id FROM folders ORDER BY name COLLATE NOCASE,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Folder{}
	for rows.Next() {
		var f Folder
		var parent sql.NullString
		if err := rows.Scan(&f.ID, &f.Name, &parent); err != nil {
			return nil, err
		}
		if parent.Valid {
			f.ParentID = &parent.String
		}
		items = append(items, f)
	}
	return items, rows.Err()
}

func (r Repository) Create(id, name string, parent *string) (Folder, error) {
	name = strings.TrimSpace(name)
	if !validName(name) {
		return Folder{}, ErrInvalid
	}
	if parent != nil {
		var exists int
		if err := r.DB.QueryRow(`SELECT 1 FROM folders WHERE id=?`, *parent).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
			return Folder{}, ErrNotFound
		} else if err != nil {
			return Folder{}, err
		}
	}
	var count int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM folders WHERE name=? COLLATE NOCASE AND parent_id IS ?`, name, parent).Scan(&count); err != nil {
		return Folder{}, err
	}
	if count > 0 {
		return Folder{}, ErrDuplicate
	}
	_, err := r.DB.Exec(`INSERT INTO folders(id,name,parent_id,created_at) VALUES(?,?,?,?)`, id, name, parent, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return Folder{}, err
	}
	return Folder{ID: id, Name: name, ParentID: parent}, nil
}

func (r Repository) Rename(id, name string) (Folder, error) {
	name = strings.TrimSpace(name)
	if !validName(name) {
		return Folder{}, ErrInvalid
	}
	var parent sql.NullString
	if err := r.DB.QueryRow(`SELECT parent_id FROM folders WHERE id=?`, id).Scan(&parent); errors.Is(err, sql.ErrNoRows) {
		return Folder{}, ErrNotFound
	} else if err != nil {
		return Folder{}, err
	}
	var duplicate int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM folders WHERE id<>? AND name=? COLLATE NOCASE AND parent_id IS ?`, id, name, nullableString(parent)).Scan(&duplicate); err != nil {
		return Folder{}, err
	}
	if duplicate > 0 {
		return Folder{}, ErrDuplicate
	}
	if _, err := r.DB.Exec(`UPDATE folders SET name=? WHERE id=?`, name, id); err != nil {
		return Folder{}, err
	}
	f := Folder{ID: id, Name: name}
	if parent.Valid {
		f.ParentID = &parent.String
	}
	return f, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func (r Repository) Move(id string, parent *string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err = tx.QueryRow(`SELECT 1 FROM folders WHERE id=?`, id).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if parent != nil {
		if err = tx.QueryRow(`SELECT 1 FROM folders WHERE id=?`, *parent).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
		var cycle int
		if err = tx.QueryRow(`WITH RECURSIVE descendants(id) AS (SELECT id FROM folders WHERE parent_id=? UNION ALL SELECT f.id FROM folders f JOIN descendants d ON f.parent_id=d.id) SELECT COUNT(*) FROM descendants WHERE id=?`, id, *parent).Scan(&cycle); err != nil {
			return err
		}
		if cycle > 0 || *parent == id {
			return ErrCycle
		}
	}
	var name string
	if err = tx.QueryRow(`SELECT name FROM folders WHERE id=?`, id).Scan(&name); err != nil {
		return err
	}
	var duplicate int
	if err = tx.QueryRow(`SELECT COUNT(*) FROM folders WHERE id<>? AND name=? COLLATE NOCASE AND parent_id IS ?`, id, name, parent).Scan(&duplicate); err != nil {
		return err
	}
	if duplicate > 0 {
		return ErrDuplicate
	}
	if _, err = tx.Exec(`UPDATE folders SET parent_id=? WHERE id=?`, parent, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r Repository) Delete(id string) error {
	var exists int
	if err := r.DB.QueryRow(`SELECT 1 FROM folders WHERE id=?`, id).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	var count int
	if err := r.DB.QueryRow(`SELECT (SELECT COUNT(*) FROM folders WHERE parent_id=?)+(SELECT COUNT(*) FROM runbooks WHERE folder_id=?)`, id, id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrNotEmpty
	}
	_, err := r.DB.Exec(`DELETE FROM folders WHERE id=?`, id)
	return err
}

func (r Repository) Valid(id *string) bool {
	if id == nil {
		return true
	}
	var n int
	return r.DB.QueryRow(`SELECT 1 FROM folders WHERE id=?`, *id).Scan(&n) == nil
}
