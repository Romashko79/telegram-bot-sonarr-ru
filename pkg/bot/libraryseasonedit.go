package bot

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/woiza/telegram-bot-sonarr/pkg/utils"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

const (
	LibrarySeasonEditToggleMonitor = "LIBRARY_SEASON_EDIT_TOGGLE_MONITOR"
	LibrarySeasonEditSubmitChanges = "LIBRARY_SEASON_EDIT_SUBMIT_CHANGES"
	LibrarySeasonEditGoBack        = "LIBRARY_SEASON_EDIT_GOBACK"
	LibrarySeasonEditCancel        = "LIBRARY_SEASON_EDIT_CANCEL"
	LibrarySeasonMonitor           = "LIBRARY_SEASON_MONITOR"
	LibrarySeasonUnmonitor         = "LIBRARY_SEASON_UNMONITOR"
	LibrarySeasonSearch            = "LIBRARY_SEASON_SEARCH"
	LibrarySeasonMonitorSearchNow  = "LIBRARY_SEASON_MONITOR_SEARCH_NOW"
	LibrarySeasonDelete            = "LIBRARY_SEASON_DELETE"
	LibrarySeasonGoBack            = "LIBRARY_SEASON_GOBACK"
)

func (b *Bot) librarySeasonEdit(update tgbotapi.Update) bool {
	chatID, err := b.getChatID(update)
	if err != nil {
		fmt.Printf("Не удалось открыть библиотеку: %v", err)
		return false
	}

	command, exists := b.getLibraryState(chatID)
	if !exists {
		return false
	}
	switch update.CallbackQuery.Data {
	case LibrarySeasonEditToggleMonitor:
		return b.handleLibrarySeasonEditToggleMonitor(command)
	case LibrarySeasonEditSubmitChanges:
		return b.handleLibrarySeasonEditSubmitChanges(update, command)
	case LibrarySeasonEditGoBack:
		b.setActiveCommand(chatID, LibraryFilteredActive)
		b.setLibraryState(command.chatID, command)
		return b.showLibrarySeriesDetail(update, command)
	case LibrarySeasonEditCancel:
		b.clearState(update)
		b.sendMessageWithEdit(command, CommandsCleared)
		return false
	case LibrarySeasonMonitor:
		return b.handleLibrarySeriesSeasonMonitor(command)
	case LibrarySeasonUnmonitor:
		return b.handleLibrarySeriesSeasonUnmonitor(command)
	case LibrarySeasonSearch:
		return b.handleLibrarySeriesSeasonSearch(command)
	case LibrarySeasonMonitorSearchNow:
		return b.handleLibrarySeriesSeasonMonitorSearchNow(command)
	case LibrarySeasonDelete:
		return b.handleLibrarySeasonDeleteSeasonUnmonitor(command)
	default:
		// Check if it starts with "SEASON_"
		if strings.HasPrefix(update.CallbackQuery.Data, "SEASON_") {
			return b.handleLibrarySeasonEditSelectSeason(update, command)
		}
		return b.showLibrarySeason(command)
	}
}

func (b *Bot) showLibrarySeason(command *userLibrary) bool {
	// Sort series seasons in ascending order
	sort.Slice(command.series.Seasons, func(i, j int) bool {
		return command.series.Seasons[i].SeasonNumber < command.series.Seasons[j].SeasonNumber
	})
	series := command.series
	messageText := fmt.Sprintf("[%v](https://www.imdb.com/title/%v) \\- _%v_\n\n", utils.Escape(series.Title), series.ImdbID, series.Year)

	var seasonKeyboardButtons [][]tgbotapi.InlineKeyboardButton
	for _, season := range command.series.Seasons {
		var buttonText string
		if season.SeasonNumber == 0 {
			buttonText = "Спецвыпуски"
		} else {
			buttonText = fmt.Sprintf("Сезон %d", season.SeasonNumber)
		}
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("SEASON_%d", season.SeasonNumber)),
		}
		seasonKeyboardButtons = append(seasonKeyboardButtons, row)
	}

	keyboardMarkup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: seasonKeyboardButtons,
	}

	keyboardGoBack := b.createKeyboard(
		[]string{"\U0001F519"},
		[]string{LibrarySeasonEditGoBack},
	)

	keyboardMarkup.InlineKeyboard = append(keyboardMarkup.InlineKeyboard, keyboardGoBack.InlineKeyboard...)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		command.chatID,
		command.messageID,
		messageText,
		keyboardMarkup,
	)
	editMsg.ParseMode = "MarkdownV2"
	editMsg.DisableWebPagePreview = true
	b.setLibraryState(command.chatID, command)
	b.sendMessage(editMsg)
	return false
}

func (b *Bot) handleLibrarySeasonEditToggleMonitor(command *userLibrary) bool {
	command.selectedMonitoring = !command.selectedMonitoring
	b.setLibraryState(command.chatID, command)
	return b.showLibrarySeriesEdit(command)
}

func (b *Bot) handleLibrarySeasonEditSelectSeason(update tgbotapi.Update, command *userLibrary) bool {
	seasonNumberStr := strings.TrimPrefix(update.CallbackQuery.Data, "SEASON_")
	seasonNumber, err := strconv.Atoi(seasonNumberStr)
	if err != nil {
		log.Println("Не удалось преобразовать номер сезона в число:", err)
		return false
	}
	command.selectedSeason = getSeasonByNumber(command.series, seasonNumber)

	b.setLibraryState(command.chatID, command)
	return b.showLibrarySeriesSeasonDetail(command)
}

func (b *Bot) showLibrarySeriesSeasonDetail(command *userLibrary) bool {
	series := command.series
	season := command.seriesSeasons[command.selectedSeason.SeasonNumber]

	var monitorIcon string
	if season.Monitored {
		monitorIcon = MonitorIcon
	} else {
		monitorIcon = UnmonitorIcon
	}

	var lastSearchString string
	if command.lastSeasonSearch[season.SeasonNumber].IsZero() {
		lastSearchString = "" // Set empty string if the time is the zero value
	} else {
		lastSearchString = command.lastSeasonSearch[season.SeasonNumber].Format("02.01.06 - 15:04") // Convert non-zero time to string
	}

	var seasonEpisodesCounter int
	for _, episode := range command.allEpisodes {
		if episode.SeasonNumber == season.SeasonNumber {
			seasonEpisodesCounter++
		}
	}

	seariesEpisodeFiles, err := b.SonarrServer.GetSeriesEpisodeFiles(command.series.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}

	var seasonEpisodeFiles []*sonarr.EpisodeFile
	var totalSize int64
	for _, file := range seariesEpisodeFiles {
		if file.SeasonNumber == season.SeasonNumber {
			totalSize += file.Size
			seasonEpisodeFiles = append(seasonEpisodeFiles, file)
		}
	}

	// Create a message with season details
	var message strings.Builder
	if season.SeasonNumber == 0 {
		fmt.Fprintf(&message, "[%v](https://www.imdb.com/title/%v) \\- _%v_ \\- Спецвыпуски\n\n", utils.Escape(series.Title), series.ImdbID, series.Year)
	} else {
		fmt.Fprintf(&message, "[%v](https://www.imdb.com/title/%v) \\- _%v_ \\- Сезон _%v_\n\n", utils.Escape(series.Title), series.ImdbID, series.Year, season.SeasonNumber)
	}
	fmt.Fprintf(&message, "Отслеживается: %s\n", monitorIcon)
	fmt.Fprintf(&message, "Последний ручной поиск: %s\n", utils.Escape(lastSearchString))
	fmt.Fprintf(&message, "Серий: %d\n", seasonEpisodesCounter)
	fmt.Fprintf(&message, "Серий на диске: %d\n", len(seasonEpisodeFiles))
	fmt.Fprintf(&message, "Размер: %d ГБ\n", totalSize/(1024*1024*1024))

	messageText := message.String()

	var keyboard tgbotapi.InlineKeyboardMarkup

	if season.Monitored && len(seasonEpisodeFiles) == 0 {
		keyboard = b.createKeyboard(
			[]string{"Не отслеживать сезон", "Искать сезон", "\U0001F519"},
			[]string{LibrarySeasonUnmonitor, LibrarySeasonSearch, LibrarySeasonGoBack},
		)
	} else if season.Monitored && len(seasonEpisodeFiles) > 0 {
		keyboard = b.createKeyboard(
			[]string{"Не отслеживать сезон", "Искать сезон", "Удалить сезон и не отслеживать", "\U0001F519"},
			[]string{LibrarySeasonUnmonitor, LibrarySeasonSearch, LibrarySeasonDelete, LibrarySeasonGoBack},
		)
	} else if !season.Monitored && len(seasonEpisodeFiles) == 0 {
		keyboard = b.createKeyboard(
			[]string{"Отслеживать сезон", "Отслеживать сезон и искать сейчас", "\U0001F519"},
			[]string{LibrarySeasonMonitor, LibrarySeasonMonitorSearchNow, LibrarySeasonGoBack},
		)
	} else if !season.Monitored && len(seasonEpisodeFiles) > 0 {
		keyboard = b.createKeyboard(
			[]string{"Отслеживать сезон", "Отслеживать сезон и искать сейчас", "Удалить сезон и не отслеживать", "\U0001F519"},
			[]string{LibrarySeasonMonitor, LibrarySeasonMonitorSearchNow, LibrarySeasonDelete, LibrarySeasonGoBack},
		)
	}
	// // Send the message containing series details along with the keyboard
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		command.chatID,
		command.messageID,
		messageText,
		keyboard,
	)
	editMsg.ParseMode = "MarkdownV2"
	editMsg.DisableWebPagePreview = true
	b.setLibraryState(command.chatID, command)
	b.sendMessage(editMsg)
	return false
}

func (b *Bot) handleLibrarySeriesSeasonMonitor(command *userLibrary) bool {
	// Update the Monitored field of the season
	command.selectedSeason.Monitored = *starr.True()
	// Convert the updated series to AddSeriesInput
	input := seriesToAddSeriesInput(command.series)
	input.Seasons[0].Monitored = *starr.True()

	// Update the series on the server
	_, err := b.SonarrServer.UpdateSeries(input, *starr.False())
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}

	b.setLibraryState(command.chatID, command)
	return b.showLibrarySeriesSeasonDetail(command)
}

func (b *Bot) handleLibrarySeriesSeasonUnmonitor(command *userLibrary) bool {
	// Access the specific season
	season := command.selectedSeason
	// Update the Monitored field of the season
	season.Monitored = *starr.False()
	// Convert the updated series to AddSeriesInput
	input := seriesToAddSeriesInput(command.series)
	// Update the series on the server
	_, err := b.SonarrServer.UpdateSeries(input, *starr.False())
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}

	b.setLibraryState(command.chatID, command)
	return b.showLibrarySeriesSeasonDetail(command)
}

func (b *Bot) handleLibrarySeriesSeasonSearch(command *userLibrary) bool {
	cmd := sonarr.CommandRequest{
		Name:         "SeasonSearch",
		SeriesID:     command.series.ID,
		SeasonNumber: command.selectedSeason.SeasonNumber,
	}
	_, err := b.SonarrServer.SendCommand(&cmd)
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}
	command.lastSeasonSearch[command.selectedSeason.SeasonNumber] = time.Now()
	b.setLibraryState(command.chatID, command)
	return b.showLibrarySeriesSeasonDetail(command)
}

func (b *Bot) handleLibrarySeriesSeasonMonitorSearchNow(command *userLibrary) bool {
	// Update the Monitored field of the season
	command.selectedSeason.Monitored = *starr.True()
	input := seriesToAddSeriesInput(command.series)
	_, err := b.SonarrServer.UpdateSeries(input, *starr.False())
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}

	cmd := sonarr.CommandRequest{
		Name:         "SeasonSearch",
		SeriesID:     command.series.ID,
		SeasonNumber: command.selectedSeason.SeasonNumber,
	}
	_, err = b.SonarrServer.SendCommand(&cmd)
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}
	command.lastSeasonSearch[command.selectedSeason.SeasonNumber] = time.Now()
	b.setLibraryState(command.chatID, command)
	return b.showLibrarySeriesSeasonDetail(command)
}

func (b *Bot) handleLibrarySeasonDeleteSeasonUnmonitor(command *userLibrary) bool {
	// Access the specific season
	season := command.selectedSeason
	episodes, err := b.SonarrServer.GetSeriesEpisodeFiles(command.series.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}
	for _, episode := range episodes {
		if episode.SeasonNumber == season.SeasonNumber {
			err := b.SonarrServer.DeleteEpisodeFile(episode.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(command.chatID, err.Error())
				b.sendMessage(msg)
				return false
			}
		}
	}

	if season.Monitored {
		// Update the Monitored field of the season
		season.Monitored = *starr.False()
		// Convert the updated series to AddSeriesInput
		input := seriesToAddSeriesInput(command.series)
		// Update the series on the server
		_, err := b.SonarrServer.UpdateSeries(input, *starr.False())
		if err != nil {
			msg := tgbotapi.NewMessage(command.chatID, err.Error())
			b.sendMessage(msg)
			return false
		}
	}

	series, err := b.SonarrServer.GetSeriesByID(command.series.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}
	command.series = series

	// get all episodeFiles
	episodeFiles, err := b.SonarrServer.GetSeriesEpisodeFiles(series.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}
	command.allEpisodeFiles = episodeFiles

	b.setLibraryState(command.chatID, command)
	return b.showLibrarySeriesSeasonDetail(command)
}

func (b *Bot) handleLibrarySeasonEditSubmitChanges(update tgbotapi.Update, command *userLibrary) bool {
	command.series.Monitored = command.selectedMonitoring
	command.series.QualityProfileID = command.selectedQualityProfile
	command.series.Tags = command.selectedTags
	input := seriesToAddSeriesInput(command.series)
	_, err := b.SonarrServer.UpdateSeries(input, *starr.False())
	if err != nil {
		msg := tgbotapi.NewMessage(command.chatID, err.Error())
		b.sendMessage(msg)
		return false
	}

	text := fmt.Sprintf("Сериал «%v» обновлён\n", command.series.Title)
	b.clearState(update)
	b.sendMessageWithEdit(command, text)
	return true
}
