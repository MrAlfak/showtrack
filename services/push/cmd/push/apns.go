package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

func sendAPNs(deviceToken string, job NotificationJob) error {
	keyPath := env("APNS_KEY_PATH", "/secrets/apns.p8")
	keyID := env("APNS_KEY_ID", "")
	teamID := env("APNS_TEAM_ID", "")
	bundleID := env("APNS_BUNDLE_ID", "")

	if _, err := os.Stat(keyPath); os.IsNotExist(err) || keyID == "" || teamID == "" || bundleID == "" {
		log.Printf("[dry-run] APNs → %s: %s", tokenPreview(deviceToken), job.Title)
		return nil
	}

	authKey, err := token.AuthKeyFromFile(keyPath)
	if err != nil {
		return err
	}

	client := apns2.NewTokenClient(&token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	})
	if env("APNS_PRODUCTION", "true") != "false" {
		client = client.Production()
	} else {
		client = client.Development()
	}

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       bundleID,
		Payload: payload.NewPayload().
			AlertTitle(job.Title).
			AlertBody(job.Body).
			Sound("default"),
	}

	res, err := client.Push(notification)
	if err != nil {
		return err
	}
	if !res.Sent() {
		return fmt.Errorf("apns rejected: %s", res.Reason)
	}
	log.Printf("APNs sent to %s", tokenPreview(deviceToken))
	return nil
}
