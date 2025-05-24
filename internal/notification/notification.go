package notification

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/models"
)

// Notifier represents the notification system
type Notifier struct {
	config *config.NotificationConfig
	db     *database.Database
	client *http.Client
}

// New creates a new notifier
func New(cfg *config.NotificationConfig, db *database.Database) *Notifier {
	return &Notifier{
		config: cfg,
		db:     db,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendPendingNotifications sends all pending notifications
func (n *Notifier) SendPendingNotifications() error {
	if !n.config.Enabled {
		return nil
	}

	// Get pending notifications
	notifications, err := n.db.GetPendingNotifications()
	if err != nil {
		return fmt.Errorf("error getting pending notifications: %w", err)
	}

	// Send each notification
	for _, notification := range notifications {
		var err error
		if notification.Type == "success" {
			err = n.sendSuccessNotification(&notification)
		} else if notification.Type == "error" {
			err = n.sendErrorNotification(&notification)
		} else {
			log.Warn().Str("type", notification.Type).Msg("Unknown notification type")
			continue
		}

		if err != nil {
			log.Error().Err(err).Int64("notification_id", notification.ID).Msg("Failed to send notification")
			continue
		}

		// Update notification status
		notification.Sent = true
		notification.SentAt = time.Now()
		if err := n.db.UpdateNotification(&notification); err != nil {
			log.Error().Err(err).Int64("notification_id", notification.ID).Msg("Failed to update notification status")
		} else {
			log.Debug().Int64("notification_id", notification.ID).Str("type", notification.Type).Msg("Notification sent successfully")
		}
	}

	return nil
}

// sendSuccessNotification sends a success notification
func (n *Notifier) sendSuccessNotification(notification *models.Notification) error {
	if n.config.Provider != "telegram" {
		return fmt.Errorf("unsupported notification provider: %s", n.config.Provider)
	}

	if n.config.TelegramToken == "" {
		return fmt.Errorf("Telegram token is not configured")
	}

	if n.config.SuccessChannel == "" {
		return fmt.Errorf("success channel is not configured")
	}

	// Send Telegram message
	return n.sendTelegramMessage(n.config.SuccessChannel, notification.Message)
}

// sendErrorNotification sends an error notification
func (n *Notifier) sendErrorNotification(notification *models.Notification) error {
	if n.config.Provider != "telegram" {
		return fmt.Errorf("unsupported notification provider: %s", n.config.Provider)
	}

	if n.config.TelegramToken == "" {
		return fmt.Errorf("Telegram token is not configured")
	}

	if n.config.ErrorGroup == "" {
		return fmt.Errorf("error group is not configured")
	}

	// Send Telegram message
	return n.sendTelegramMessage(n.config.ErrorGroup, notification.Message)
}

// sendTelegramMessage sends a message to a Telegram chat
func (n *Notifier) sendTelegramMessage(chatID, message string) error {
	// Build the URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.config.TelegramToken)

	// Build the request
	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", message)
	params.Set("parse_mode", "Markdown")

	// Send the request
	resp, err := n.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("error sending Telegram message: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending Telegram message: %s", resp.Status)
	}

	return nil
}
