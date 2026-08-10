package push

import (
	"context"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FirebaseSender struct {
	client *messaging.Client
}

func NewFirebaseSender(ctx context.Context, projectID string, serviceAccountFile string) (*FirebaseSender, error) {
	serviceAccountFile = strings.TrimSpace(serviceAccountFile)
	if serviceAccountFile == "" {
		return nil, fmt.Errorf("firebase service account file is required")
	}

	config := &firebase.Config{}
	if strings.TrimSpace(projectID) != "" {
		config.ProjectID = strings.TrimSpace(projectID)
	}

	app, err := firebase.NewApp(ctx, config, option.WithCredentialsFile(serviceAccountFile))
	if err != nil {
		return nil, fmt.Errorf("create firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("create firebase messaging client: %w", err)
	}

	return &FirebaseSender{client: client}, nil
}

func (s *FirebaseSender) Send(ctx context.Context, message Message) error {
	if strings.TrimSpace(message.Token) == "" {
		return fmt.Errorf("push token is required")
	}

	firebaseMessage := &messaging.Message{
		Token: message.Token,
		Notification: &messaging.Notification{
			Title: message.Title,
			Body:  message.Body,
		},
		Data: message.Data,
	}
	androidConfig := &messaging.AndroidConfig{
		Priority: "normal",
		Notification: &messaging.AndroidNotification{
			ChannelID: "reminders",
		},
	}
	if message.Importance == ImportanceUrgent {
		androidConfig.Priority = "high"
		androidConfig.Notification.ChannelID = "urgent_reminders"
	}
	firebaseMessage.Android = androidConfig

	_, err := s.client.Send(ctx, firebaseMessage)
	if err != nil {
		return fmt.Errorf("send firebase message: %w", err)
	}

	return nil
}
