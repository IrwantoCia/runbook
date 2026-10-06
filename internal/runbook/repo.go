package runbook

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrNotFound = errors.New("runbook not found")

type Repository struct{ DB *sql.DB }

func (r Repository) Create(id, slug, title, body, creator string, folderID *string) (Runbook, error) {
	now := time.Now().UTC()
	_, err := r.DB.Exec(`INSERT INTO runbooks(id,slug,title,body,status,created_by,created_at,updated_at,folder_id) VALUES(?,?,?,?,?,?,?,?,?)`, id, slug, title, body, Draft, creator, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), folderID)
	if err != nil {
		return Runbook{}, err
	}
	return r.Get(id)
}

func (r Repository) Get(id string) (Runbook, error) {
	var item Runbook
	var created, updated string
	var published sql.NullString
	var folderID sql.NullString
	err := r.DB.QueryRow(`SELECT id,slug,title,body,status,created_by,created_at,updated_at,published_at,folder_id FROM runbooks WHERE id=?`, id).Scan(&item.ID, &item.Slug, &item.Title, &item.Body, &item.Status, &item.CreatedBy, &created, &updated, &published, &folderID)
	if errors.Is(err, sql.ErrNoRows) {
		return item, ErrNotFound
	}
	if err != nil {
		return item, err
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	if published.Valid {
		value, _ := time.Parse(time.RFC3339Nano, published.String)
		item.PublishedAt = &value
	}
	if folderID.Valid {
		item.FolderID = &folderID.String
	}
	return item, nil
}

func (r Repository) List(status string, publicOnly bool, folderID *string) ([]Runbook, error) {
	query := `SELECT id,slug,title,body,status,created_by,created_at,updated_at,published_at,folder_id FROM runbooks WHERE 1=1`
	args := []any{}
	if publicOnly {
		query += ` AND status='published'`
	} else if status != "" {
		query += ` AND status=?`
		args = append(args, status)
	}
	if !publicOnly {
		query += ` AND folder_id IS ?`
		args = append(args, folderID)
	}
	query += ` ORDER BY updated_at DESC`
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Runbook{}
	for rows.Next() {
		var item Runbook
		var created, updated string
		var published sql.NullString
		var folderID sql.NullString
		if err := rows.Scan(&item.ID, &item.Slug, &item.Title, &item.Body, &item.Status, &item.CreatedBy, &created, &updated, &published, &folderID); err != nil {
			return nil, err
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		if published.Valid {
			value, _ := time.Parse(time.RFC3339Nano, published.String)
			item.PublishedAt = &value
		}
		if folderID.Valid {
			item.FolderID = &folderID.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Repository) Move(id string, folderID *string) (Runbook, error) {
	result, err := r.DB.Exec(`UPDATE runbooks SET folder_id=?,updated_at=? WHERE id=?`, folderID, time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return Runbook{}, err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return Runbook{}, ErrNotFound
	}
	return r.Get(id)
}

func (r Repository) Update(id string, title, body *string) (Runbook, error) {
	current, err := r.Get(id)
	if err != nil {
		return current, err
	}
	if current.Status == Published {
		return current, errors.New("published runbooks are immutable; reject to draft first")
	}
	if title != nil {
		current.Title = strings.TrimSpace(*title)
	}
	if body != nil {
		current.Body = *body
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = r.DB.Exec(`UPDATE runbooks SET title=?,body=?,updated_at=? WHERE id=?`, current.Title, current.Body, now, id)
	if err != nil {
		return current, err
	}
	return r.Get(id)
}

func (r Repository) Transition(id string, from, to Status) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var result sql.Result
	var err error
	if to == Published {
		result, err = r.DB.Exec(`UPDATE runbooks SET status=?,published_at=?,updated_at=? WHERE id=? AND status=?`, to, now, now, id, from)
	} else {
		result, err = r.DB.Exec(`UPDATE runbooks SET status=?,published_at=NULL,updated_at=? WHERE id=? AND status=?`, to, now, id, from)
	}
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		if _, err := r.Get(id); err != nil {
			return err
		}
		return errors.New("invalid runbook status transition")
	}
	return nil
}

func (r Repository) Delete(id string) error {
	result, err := r.DB.Exec(`DELETE FROM runbooks WHERE id=?`, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r Repository) PublicBySlug(slug string) (Runbook, error) {
	var id string
	err := r.DB.QueryRow(`SELECT id FROM runbooks WHERE slug=? AND status='published'`, slug).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return Runbook{}, ErrNotFound
	}
	if err != nil {
		return Runbook{}, err
	}
	return r.Get(id)
}
