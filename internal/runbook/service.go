package runbook

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

var slugChars = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(title string) string {
	slug := strings.Trim(slugChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(title)), "-"), "-")
	if slug == "" {
		return "runbook"
	}
	return slug
}
func newID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type Service struct{ Repo Repository }

func (s Service) Create(title, body, creator string, folderID *string) (Runbook, error) {
	id, err := newID()
	if err != nil {
		return Runbook{}, err
	}
	base := Slugify(title)
	slug := base
	for suffix := 2; ; suffix++ {
		var count int
		err := s.Repo.DB.QueryRow(`SELECT COUNT(*) FROM runbooks WHERE slug=?`, slug).Scan(&count)
		if err != nil {
			return Runbook{}, err
		}
		if count == 0 {
			break
		}
		slug = base + "-" + itoa(suffix)
	}
	return s.Repo.Create(id, slug, strings.TrimSpace(title), body, creator, folderID)
}
func itoa(value int) string               { return strconv.Itoa(value) }
func (s Service) Publish(id string) error { return s.Repo.Transition(id, Draft, Published) }
