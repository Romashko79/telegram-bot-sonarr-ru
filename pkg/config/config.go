package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// BotConfig ...
type Config struct {
	TelegramBotToken string
	AllowedChatIDs   map[int64]bool
	MaxItems         int
	IgnoreTags       bool
	SeriesType       string
	SonarrProtocol   string
	SonarrHostname   string
	SonarrPort       int
	SonarrAPIKey     string
	SonarrBaseUrl    string
	WebhookPort      int
	WebhookToken     string
}

func LoadConfig() (Config, error) {
	var config Config

	config.TelegramBotToken = os.Getenv("SBOT_TELEGRAM_BOT_TOKEN")
	allowedUserIDs := os.Getenv("SBOT_BOT_ALLOWED_USERIDS")
	botMaxItems := os.Getenv("SBOT_BOT_MAX_ITEMS")
	botIgnoreTags := os.Getenv("SBOT_BOT_IGNORE_TAGS")
	botSeriesType := os.Getenv("SBOT_BOT_SERIES_TYPE")
	config.SonarrProtocol = os.Getenv("SBOT_SONARR_PROTOCOL")
	config.SonarrHostname = os.Getenv("SBOT_SONARR_HOSTNAME")
	sonarrPort := os.Getenv("SBOT_SONARR_PORT")
	config.SonarrAPIKey = os.Getenv("SBOT_SONARR_API_KEY")
	config.SonarrBaseUrl = os.Getenv("SBOT_SONARR_BASE_URL")

	// Validate required fields
	if config.TelegramBotToken == "" {
		return config, errors.New("Переменная SBOT_TELEGRAM_BOT_TOKEN пуста или не задана")
	}
	if allowedUserIDs == "" {
		return config, errors.New("Переменная SBOT_BOT_ALLOWED_USERIDS пуста или не задана")
	}
	if botMaxItems == "" {
		return config, errors.New("Переменная SBOT_BOT_MAX_ITEMS пуста или не задана")
	}
	if botIgnoreTags == "" {
		return config, errors.New("Переменная SBOT_BOT_IGNORE_TAGS пуста или не задана")
	}
	// Normalize and validate SBOT_SONARR_PROTOCOL
	config.SonarrProtocol = strings.ToLower(config.SonarrProtocol)
	if config.SonarrProtocol != "http" && config.SonarrProtocol != "https" {
		return config, errors.New("Переменная SBOT_SONARR_PROTOCOL должна быть http или https")
	}
	if config.SonarrHostname == "" {
		return config, errors.New("Переменная SBOT_SONARR_HOSTNAME пуста или не задана")
	}
	if sonarrPort == "" {
		return config, errors.New("Переменная SBOT_SONARR_PORT пуста или не задана")
	}
	if config.SonarrAPIKey == "" {
		return config, errors.New("Переменная SBOT_SONARR_API_KEY пуста или не задана")
	}

	// Parsing SBOT_BOT_MAX_ITEMS as a number
	maxItems, err := strconv.Atoi(botMaxItems)
	if err != nil {
		return config, errors.New("Переменная SBOT_BOT_MAX_ITEMS должна быть числом")
	}
	config.MaxItems = maxItems

	// Parsing SBOT_BOT_IGNORE_TAGS as a boolean
	ignoreTags, err := strconv.ParseBool(botIgnoreTags)
	if err != nil {
		return config, errors.New("Переменная SBOT_BOT_IGNORE_TAGS должна быть true или false")
	}
	config.IgnoreTags = ignoreTags

	// Normalize and validate SBOT_BOT_SERIES_TYPE
	if strings.EqualFold(botSeriesType, "standard") {
		config.SeriesType = "standard"
	} else if strings.EqualFold(botSeriesType, "daily") {
		config.SeriesType = "daily"
	} else if strings.EqualFold(botSeriesType, "anime") {
		config.SeriesType = "anime"
	} else {
		config.SeriesType = ""
	}

	// Parsing SBOT_BOT_ALLOWED_USERIDS as a list of integers
	userIDs := strings.Split(allowedUserIDs, ",")
	parsedUserIDs := make(map[int64]bool)
	for _, id := range userIDs {
		parsedID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return config, fmt.Errorf("Переменная SBOT_BOT_ALLOWED_USERIDS содержит нецелое значение: %s", err)
		}
		parsedUserIDs[parsedID] = true
	}
	config.AllowedChatIDs = parsedUserIDs

	// Parsing SBOT_SONARR_PORT as a number
	port, err := strconv.Atoi(sonarrPort)
	if err != nil {
		return config, errors.New("Переменная SBOT_SONARR_PORT должна быть числом")
	}
	config.SonarrPort = port

	// Optional: receiver for Sonarr webhook notifications
	config.WebhookToken = os.Getenv("SBOT_WEBHOOK_TOKEN")
	if webhookPort := os.Getenv("SBOT_WEBHOOK_PORT"); webhookPort != "" {
		wp, err := strconv.Atoi(webhookPort)
		if err != nil {
			return config, errors.New("Переменная SBOT_WEBHOOK_PORT должна быть числом")
		}
		config.WebhookPort = wp
	}

	return config, nil
}
