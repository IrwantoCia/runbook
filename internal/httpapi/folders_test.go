package httpapi

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"runbook/internal/auth"
	"runbook/internal/config"
	"runbook/internal/db"
)

func TestFolderEndpointsRequireSession(t *testing.T) {
	database := testHTTPDatabase(t)
	router := NewRouter(database, config.Config{SessionTTLHours: 1})
	for _, request := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/folders", ""},
		{http.MethodPost, "/api/folders", `{"name":"Folder"}`},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(request.method, request.path, strings.NewReader(request.body)))
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", request.method, request.path, response.Code)
		}
	}
	if _, err := database.Exec(`INSERT INTO sessions(id,user_id,token_hash,expires_at,created_at) VALUES('s1','u1',?,?,?)`, auth.HashToken("session-token"), time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/folders", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: "session-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("authenticated GET /api/folders status = %d, body=%s", response.Code, response.Body.String())
	}
	for _, test := range []struct {
		body string
		want int
	}{
		{`{"name":""}`, http.StatusBadRequest},
		{`{"name":"Missing parent","parent_id":"absent"}`, http.StatusNotFound},
	} {
		request = httptest.NewRequest(http.MethodPost, "/api/folders", strings.NewReader(test.body))
		request.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: "session-token"})
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != test.want {
			t.Errorf("POST /api/folders %s status = %d, want %d", test.body, response.Code, test.want)
		}
	}
}

func TestSecurityHeadersOnAPIResponse(t *testing.T) {
	database := testHTTPDatabase(t)
	router := NewRouter(database, config.Config{CookieSecure: true})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	for name, want := range map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Referrer-Policy":           "no-referrer",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	} {
		if got := response.Header().Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestLoginBodyLimitReturnsBadRequest(t *testing.T) {
	database := testHTTPDatabase(t)
	router := NewRouter(database, config.Config{})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(strings.Repeat("x", (2<<20)+1))))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("oversized login status = %d, want 400; body=%s", response.Code, response.Body.String())
	}
}

func TestLoginFailuresReturnUnauthorizedWithoutRateLimit(t *testing.T) {
	database := testHTTPDatabase(t)
	passwordHash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE users SET password_hash=? WHERE id='u1'`, passwordHash); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(database, config.Config{})
	for i := 0; i < 10; i++ {
		body := `{"email":"missing@example.com","password":"wrong-password"}`
		if i%2 == 1 {
			body = `{"email":"owner@example.com","password":"wrong-password"}`
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body)))
		if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "invalid email or password") {
			t.Fatalf("login attempt %d status=%d body=%s, want 401 invalid credentials", i+1, response.Code, response.Body.String())
		}
	}
}

func TestRunbookListQueryDistinguishesRootFromFolder(t *testing.T) {
	database := testHTTPDatabase(t)
	_, err := database.Exec(`
INSERT INTO folders(id,name,created_at) VALUES('f1','Folder','now');
INSERT INTO runbooks(id,slug,title,body,status,created_by,created_at,updated_at,folder_id) VALUES
 ('root-id','root-book','Root','root','draft','u1','now','now',NULL),
 ('folder-id','folder-book','Nested','nested','draft','u1','now','now','f1');
INSERT INTO sessions(id,user_id,token_hash,expires_at,created_at) VALUES('s1','u1',?,?,?)`, auth.HashToken("session-token"), time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(database, config.Config{SessionTTLHours: 1})
	for _, test := range []struct {
		path, want, exclude string
	}{
		{"/api/runbooks", "root-book", "folder-book"},
		{"/api/runbooks?folder_id=f1", "folder-book", "root-book"},
	} {
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		request.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: "session-token"})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.want) || strings.Contains(response.Body.String(), test.exclude) {
			t.Errorf("GET %s status=%d body=%s; want %q and exclude %q", test.path, response.Code, response.Body.String(), test.want, test.exclude)
		}
	}
}

func TestMachineAPIExposesOnlyPublishedRunbooks(t *testing.T) {
	database := testHTTPDatabase(t)
	if _, err := database.Exec(`
INSERT INTO api_keys(id,name,key_hash,prefix,created_at) VALUES('k1','test key',?,'prefix','now');
INSERT INTO runbooks(id,slug,title,body,status,created_by,created_at,updated_at) VALUES
 ('draft-id','private-draft','Private title','PRIVATE BODY','draft','u1','now','now'),
 ('published-id','public-slug','Public title','PUBLIC BODY','published','u1','now','now');`, auth.HashToken("machine-token")); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(database, config.Config{SessionTTLHours: 1})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/runbooks", nil)
	request.Header.Set("Authorization", "Bearer machine-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("public list status = %d, body=%s", response.Code, response.Body.String())
	}
	if body := response.Body.String(); strings.Contains(body, "private-draft") || strings.Contains(body, "PRIVATE BODY") || !strings.Contains(body, "public-slug") {
		t.Fatalf("public list leaked or omitted runbook: %s", body)
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/runbooks/private-draft", nil)
	request.Header.Set("Authorization", "Bearer machine-token")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("draft slug status = %d, want 404; body=%s", response.Code, response.Body.String())
	}
}

func testHTTPDatabase(t *testing.T) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(id,email,password_hash,created_at) VALUES('u1','owner@example.com','hash','now')`); err != nil {
		t.Fatal(err)
	}
	return database
}
