package api

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-server/utils"
)

const (
	maxBodyBytes       = 1 << 20
	passwordIterations = 600_000
	sessionDuration    = 30 * 24 * time.Hour
)

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)
var reservedSlugs = map[string]bool{"api": true, "dashboard": true, "go": true, "healthz": true, "login": true, "signup": true, "static": true}

type App struct {
	DB          *sql.DB
	Now         func() time.Time
	LinkBaseURL string
}
type link struct {
	Slug      string     `json:"slug"`
	ShortURL  string     `json:"short_url"`
	URL       string     `json:"url"`
	OwnerID   *int       `json:"owner_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
	Clicks    int        `json:"clicks"`
}
type createLinkRequest struct {
	Slug      string     `json:"slug"`
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expires_at"`
}
type createUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type updateUserRequest struct {
	Username string `json:"username"`
}

func New(db *sql.DB, linkBaseURL string) *App {
	return &App{DB: db, Now: time.Now, LinkBaseURL: linkBaseURL}
}

func (app *App) Health(w http.ResponseWriter, r *http.Request) {
	if err := app.DB.PingContext(r.Context()); err != nil {
		utils.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *App) Links(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		userID, ok := app.requireUser(w, r)
		if !ok {
			return
		}
		links, err := app.findLinksByOwner(r, userID)
		if err != nil {
			log.Printf("list links: %v", err)
			utils.WriteError(w, http.StatusInternalServerError, "could not retrieve links")
			return
		}
		utils.WriteJSON(w, http.StatusOK, links)
		return
	}
	if r.Method != http.MethodPost {
		utils.MethodNotAllowed(w, http.MethodGet, http.MethodPost)
		return
	}
	var input createLinkRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := validateLinkInput(&input); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Slug == "" {
		var err error
		input.Slug, err = newSlug()
		if err != nil {
			utils.WriteError(w, 500, "could not generate a slug")
			return
		}
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(app.Now()) {
		utils.WriteError(w, 400, "expires_at must be in the future")
		return
	}
	userID, ok := app.requireUser(w, r)
	if !ok {
		return
	}
	_, err := app.DB.ExecContext(r.Context(), `INSERT INTO Short (slug, url, ownerId, expiresAt) VALUES (?, ?, ?, ?)`, input.Slug, input.URL, userID, input.ExpiresAt)
	if err != nil {
		if isDuplicate(err) {
			utils.WriteError(w, 409, "slug is already in use")
		} else {
			log.Printf("create link: %v", err)
			utils.WriteError(w, 500, "could not create link")
		}
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]string{"slug": input.Slug, "short_url": app.ShortURL(r, input.Slug)})
}

func (app *App) Link(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/links/")
	if !slugPattern.MatchString(slug) || strings.Contains(slug, "/") {
		utils.WriteError(w, 400, "invalid slug")
		return
	}
	userID, ok := app.requireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		link, err := app.findOwnedLink(r, slug, userID)
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, 404, "link not found")
			return
		}
		if err != nil {
			log.Printf("get link: %v", err)
			utils.WriteError(w, 500, "could not retrieve link")
			return
		}
		utils.WriteJSON(w, 200, link)
	case http.MethodDelete:
		result, err := app.DB.ExecContext(r.Context(), `DELETE FROM Short WHERE slug = ? AND ownerId = ?`, slug, userID)
		if err != nil {
			log.Printf("delete link: %v", err)
			utils.WriteError(w, 500, "could not delete link")
			return
		}
		count, err := result.RowsAffected()
		if err != nil || count == 0 {
			utils.WriteError(w, 404, "link not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		utils.MethodNotAllowed(w, http.MethodGet, http.MethodDelete)
	}
}

func (app *App) Redirect(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !slugPattern.MatchString(slug) {
		http.NotFound(w, r)
		return
	}
	link, err := app.findLink(r, slug)
	if errors.Is(err, sql.ErrNoRows) || err == nil && (!link.IsActive || link.ExpiresAt != nil && !link.ExpiresAt.After(app.Now())) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("redirect lookup: %v", err)
		utils.WriteError(w, 500, "could not resolve link")
		return
	}
	if _, err := app.DB.ExecContext(r.Context(), `UPDATE Short SET clicks = clicks + 1 WHERE slug = ?`, slug); err != nil {
		log.Printf("increment click: %v", err)
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}

func (app *App) Users(w http.ResponseWriter, r *http.Request) {
	var input createUserRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if len(input.Username) < 3 || len(input.Username) > 96 {
		utils.WriteError(w, 400, "username must be 3 to 96 characters")
		return
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		utils.WriteError(w, 400, "email is invalid")
		return
	}
	if len(input.Password) < 12 {
		utils.WriteError(w, 400, "password must be at least 12 characters")
		return
	}
	hash, err := hashPassword(input.Password)
	if err != nil {
		utils.WriteError(w, 500, "could not create user")
		return
	}
	result, err := app.DB.ExecContext(r.Context(), `INSERT INTO User (username, password, email) VALUES (?, ?, ?)`, input.Username, hash, input.Email)
	if err != nil {
		if isDuplicate(err) {
			utils.WriteError(w, 409, "username or email is already in use")
		} else {
			log.Printf("create user: %v", err)
			utils.WriteError(w, 500, "could not create user")
		}
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		utils.WriteError(w, 500, "could not create user")
		return
	}
	if err := app.startSession(w, r, id); err != nil {
		log.Printf("create session: %v", err)
		utils.WriteError(w, 500, "could not create session")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// Me returns or updates the signed-in account's username.
func (app *App) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := app.requireUser(w, r)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		var username string
		err := app.DB.QueryRowContext(r.Context(), `SELECT username FROM User WHERE id = ?`, userID).Scan(&username)
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		if err != nil {
			log.Printf("get user: %v", err)
			utils.WriteError(w, http.StatusInternalServerError, "could not retrieve user")
			return
		}
		utils.WriteJSON(w, http.StatusOK, map[string]string{"username": username})
		return
	}
	if r.Method != http.MethodPut {
		utils.MethodNotAllowed(w, http.MethodGet, http.MethodPut)
		return
	}
	var input updateUserRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if len(input.Username) < 3 || len(input.Username) > 96 {
		utils.WriteError(w, http.StatusBadRequest, "username must be 3 to 96 characters")
		return
	}
	if _, err := app.DB.ExecContext(r.Context(), `UPDATE User SET username = ? WHERE id = ?`, input.Username, userID); err != nil {
		if isDuplicate(err) {
			utils.WriteError(w, http.StatusConflict, "username is already in use")
		} else {
			log.Printf("update user: %v", err)
			utils.WriteError(w, http.StatusInternalServerError, "could not update user")
		}
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"username": input.Username})
}

func (app *App) Sessions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		app.endSession(w, r)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var input loginRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Username == "" || input.Password == "" {
		utils.WriteError(w, 400, "username and password are required")
		return
	}
	var id int64
	var encoded string
	err := app.DB.QueryRowContext(r.Context(), `SELECT id, password FROM User WHERE username = ?`, input.Username).Scan(&id, &encoded)
	if errors.Is(err, sql.ErrNoRows) || err == nil && !verifyPassword(encoded, input.Password) {
		utils.WriteError(w, 401, "invalid username or password")
		return
	}
	if err != nil {
		log.Printf("login: %v", err)
		utils.WriteError(w, 500, "could not authenticate")
		return
	}
	if err := app.startSession(w, r, id); err != nil {
		log.Printf("create session: %v", err)
		utils.WriteError(w, 500, "could not create session")
		return
	}
	utils.WriteJSON(w, 200, map[string]int64{"id": id})
}

func (app *App) DeleteExpiredLinks(ctx context.Context) {
	_, _ = app.DB.ExecContext(ctx, `DELETE FROM Short WHERE expiresAt IS NOT NULL AND expiresAt <= ?`, app.Now())
}
func (app *App) CleanupExpiredLinks() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		app.DeleteExpiredLinks(context.Background())
	}
}
func (app *App) ShortURL(r *http.Request, slug string) string {
	if app.LinkBaseURL != "" {
		return app.LinkBaseURL + "/" + slug
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/go/" + slug
}

func (app *App) findLink(r *http.Request, slug string) (link, error) {
	var value link
	err := app.DB.QueryRowContext(r.Context(), `SELECT slug, url, ownerId, createdAt, expiresAt, isActive, clicks FROM Short WHERE slug = ?`, slug).Scan(&value.Slug, &value.URL, &value.OwnerID, &value.CreatedAt, &value.ExpiresAt, &value.IsActive, &value.Clicks)
	return value, err
}
func (app *App) findOwnedLink(r *http.Request, slug string, userID int64) (link, error) {
	var value link
	err := app.DB.QueryRowContext(r.Context(), `SELECT slug, url, ownerId, createdAt, expiresAt, isActive, clicks FROM Short WHERE slug = ? AND ownerId = ?`, slug, userID).Scan(&value.Slug, &value.URL, &value.OwnerID, &value.CreatedAt, &value.ExpiresAt, &value.IsActive, &value.Clicks)
	if err == nil {
		value.ShortURL = app.ShortURL(r, value.Slug)
	}
	return value, err
}
func (app *App) findLinksByOwner(r *http.Request, userID int64) ([]link, error) {
	rows, err := app.DB.QueryContext(r.Context(), `SELECT slug, url, ownerId, createdAt, expiresAt, isActive, clicks FROM Short WHERE ownerId = ? ORDER BY createdAt DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := []link{}
	for rows.Next() {
		var value link
		if err := rows.Scan(&value.Slug, &value.URL, &value.OwnerID, &value.CreatedAt, &value.ExpiresAt, &value.IsActive, &value.Clicks); err != nil {
			return nil, err
		}
		value.ShortURL = app.ShortURL(r, value.Slug)
		links = append(links, value)
	}
	return links, rows.Err()
}
func (app *App) requireUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	cookie, err := r.Cookie("golink_session")
	if err != nil || cookie.Value == "" {
		utils.WriteError(w, 401, "sign in to manage links")
		return 0, false
	}
	var userID int64
	err = app.DB.QueryRowContext(r.Context(), `SELECT userId FROM Session WHERE tokenHash = ? AND expiresAt > ?`, sessionTokenHash(cookie.Value), app.Now()).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		utils.WriteError(w, 401, "sign in to manage links")
		return 0, false
	}
	if err != nil {
		log.Printf("authenticate session: %v", err)
		utils.WriteError(w, 500, "could not authenticate session")
		return 0, false
	}
	return userID, true
}
func (app *App) startSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, err := newSessionToken()
	if err != nil {
		return err
	}
	expiresAt := app.Now().Add(sessionDuration)
	if _, err := app.DB.ExecContext(r.Context(), `INSERT INTO Session (tokenHash, userId, expiresAt) VALUES (?, ?, ?)`, sessionTokenHash(token), userID, expiresAt); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "golink_session", Value: token, Path: "/", Expires: expiresAt, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil})
	return nil
}
func (app *App) endSession(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("golink_session"); err == nil {
		_, _ = app.DB.ExecContext(r.Context(), `DELETE FROM Session WHERE tokenHash = ?`, sessionTokenHash(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: "golink_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil})
}
func validateLinkInput(input *createLinkRequest) error {
	if input.URL == "" {
		return errors.New("url is required")
	}
	u, err := url.ParseRequestURI(input.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("url must be an absolute http or https URL")
	}
	if input.Slug != "" && !slugPattern.MatchString(input.Slug) {
		return errors.New("slug must be 3 to 64 letters, numbers, underscores, or hyphens")
	}
	if reservedSlugs[strings.ToLower(input.Slug)] {
		return errors.New("slug is reserved")
	}
	return nil
}
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		utils.WriteError(w, 400, "request body must be valid JSON")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		utils.WriteError(w, 400, "request body must contain one JSON object")
		return false
	}
	return true
}
func newSlug() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", passwordIterations, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}
func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 1 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(want))
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}
func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func sessionTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
func isDuplicate(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique constraint")
}
