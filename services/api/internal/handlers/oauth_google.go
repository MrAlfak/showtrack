package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type googleAuthRequest struct {
	IDToken string `json:"id_token"`
}

func (h *Handler) GoogleAuth(c *fiber.Ctx) error {
	if h.cfg.GoogleClientID == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "google oauth not configured")
	}

	var req googleAuthRequest
	if err := c.BodyParser(&req); err != nil || req.IDToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id_token required")
	}

	profile, err := verifyGoogleIDToken(req.IDToken, h.cfg.GoogleClientID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid google token")
	}
	if profile.Email == "" || profile.Sub == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "google profile incomplete")
	}

	isNew := false
	var userID uuid.UUID
	err = h.db.QueryRow(context.Background(),
		`SELECT id FROM users WHERE google_sub = $1`, profile.Sub,
	).Scan(&userID)
	if err == pgx.ErrNoRows {
		err = h.db.QueryRow(context.Background(),
			`SELECT id FROM users WHERE email = $1`, profile.Email,
		).Scan(&userID)
		if err == nil {
			_, err = h.db.Exec(context.Background(),
				`UPDATE users SET google_sub = $1, display_name = COALESCE(display_name, $2), avatar_url = COALESCE(avatar_url, $3) WHERE id = $4`,
				profile.Sub, profile.Name, profile.Picture, userID,
			)
		} else if err == pgx.ErrNoRows {
			isNew = true
			userID = uuid.New()
			_, err = h.db.Exec(context.Background(), `
				INSERT INTO users (id, email, password_hash, display_name, google_sub, avatar_url)
				VALUES ($1, $2, NULL, $3, $4, $5)`,
				userID, profile.Email, profile.Name, profile.Sub, profile.Picture,
			)
		}
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	token, err := h.signToken(userID.String())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "token failed")
	}
	return c.JSON(fiber.Map{"token": token, "user_id": userID, "is_new": isNew})
}

type googleProfile struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

func verifyGoogleIDToken(idToken, clientID string) (*googleProfile, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("tokeninfo status %d", res.StatusCode)
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if unmarshalErr := json.Unmarshal(raw, &payload); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	aud, _ := payload["aud"].(string)
	if aud != clientID {
		return nil, fmt.Errorf("audience mismatch")
	}
	emailVerified := payload["email_verified"]
	if emailVerified == false || emailVerified == "false" {
		return nil, fmt.Errorf("email not verified")
	}
	return &googleProfile{
		Sub:     fmt.Sprint(payload["sub"]),
		Email:   fmt.Sprint(payload["email"]),
		Name:    fmt.Sprint(payload["name"]),
		Picture: fmt.Sprint(payload["picture"]),
	}, nil
}
