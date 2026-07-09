package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2/google"
)

type fcmSender struct {
	projectID   string
	tokenSource func(context.Context) (string, error)
}

var (
	fcmOnce   sync.Once
	fcmClient *fcmSender
)

func getFCM() *fcmSender {
	fcmOnce.Do(func() {
		saPath := env("FCM_SERVICE_ACCOUNT", "/secrets/fcm.json")
		data, err := os.ReadFile(saPath)
		if err != nil {
			return
		}
		var meta struct {
			ProjectID string `json:"project_id"`
		}
		if err := json.Unmarshal(data, &meta); err != nil || meta.ProjectID == "" {
			log.Printf("fcm: invalid service account json")
			return
		}
		creds, err := google.CredentialsFromJSON(context.Background(), data, "https://www.googleapis.com/auth/firebase.messaging")
		if err != nil {
			log.Printf("fcm: credentials error: %v", err)
			return
		}
		fcmClient = &fcmSender{
			projectID: meta.ProjectID,
			tokenSource: func(ctx context.Context) (string, error) {
				token, err := creds.TokenSource.Token()
				if err != nil {
					return "", err
				}
				return token.AccessToken, nil
			},
		}
	})
	return fcmClient
}

func sendFCM(token string, job NotificationJob) error {
	client := getFCM()
	if client == nil {
		log.Printf("[dry-run] FCM → %s: %s", tokenPreview(token), job.Title)
		return nil
	}

	accessToken, err := client.tokenSource(context.Background())
	if err != nil {
		return err
	}

	payload := map[string]any{
		"message": map[string]any{
			"token": token,
			"notification": map[string]string{
				"title": job.Title,
				"body":  job.Body,
			},
			"data": map[string]string{
				"show_id":    fmt.Sprint(job.ShowID),
				"episode_id": fmt.Sprint(job.Episode),
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", client.projectID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		response, _ := io.ReadAll(res.Body)
		return fmt.Errorf("fcm status %d: %s", res.StatusCode, string(response))
	}
	log.Printf("FCM sent to %s", tokenPreview(token))
	return nil
}

func tokenPreview(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:8] + "..."
}
