package bot

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/woiza/telegram-bot-sonarr/pkg/utils"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

func (b *Bot) handleCommand(update tgbotapi.Update, s *sonarr.Sonarr) {

	chatID, err := b.getChatID(update)
	if err != nil {
		fmt.Printf("Не удалось обработать команду: %v", err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "")

	switch update.Message.Command() {

	case "q", "query", "add", "Q", "Query", "Add":
		b.setActiveCommand(chatID, AddSeriesCommand)
		b.processAddCommand(update, chatID, s)

	case "series", "library", "l":
		b.setActiveCommand(chatID, LibraryMenuCommand)
		b.processLibraryCommand(update, chatID, s)

	case "delete", "remove", "Delete", "Remove", "d":
		b.setActiveCommand(chatID, DeleteSeriesCommand)
		b.processDeleteCommand(update, chatID, s)

	case "clear", "cancel", "stop":
		b.clearState(update)
		msg.Text = "Все команды сброшены"
		b.sendMessage(msg)

	case "diskspace", "disk", "free", "rootfolder", "rootfolders":
		rootFolders, err := s.GetRootFolders()
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = utils.PrepareRootFolders(rootFolders)
		msg.ParseMode = "MarkdownV2"
		msg.DisableWebPagePreview = true
		b.sendMessage(msg)

	case "up", "upcoming":
		calendar := sonarr.Calendar{
			Start:       time.Now(),
			End:         time.Now().AddDate(0, 0, 30), // 30 days
			Unmonitored: *starr.True(),
		}
		upcoming, err := s.GetCalendar(calendar)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		if len(upcoming) == 0 {
			msg.Text = "В ближайшие 30 дней релизов нет"
			b.sendMessage(msg)
			break
		}
		b.sendUpcoming(upcoming, &msg)

	case "rss", "RSS":
		command := sonarr.CommandRequest{
			Name:      "RssSync",
			SeriesIDs: []int64{},
		}
		_, err := s.SendCommand(&command)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = "RSS-синхронизация запущена"
		b.sendMessage(msg)

	// does not work
	// case "searchmonitored":
	// 	series, err := s.GetSeries(0)
	// 	if err != nil {
	// 		msg.Text = err.Error()
	// 		fmt.Println(err)
	// 		b.sendMessage(msg)
	// 		break
	// 	}
	// 	var monitoredSeriesIDs []int64
	// 	for _, series := range series {
	// 		if series.Monitored {
	// 			monitoredSeriesIDs = append(monitoredSeriesIDs, series.ID)
	// 		}
	// 	}
	// 	command := sonarr.CommandRequest{
	// 		Name:      "SeriesSearch",
	// 		SeriesIDs: monitoredSeriesIDs,
	// 	}
	// 	_, err = s.SendCommand(&command)
	// 	if err != nil {
	// 		msg.Text = err.Error()
	// 		fmt.Println(err)
	// 		b.sendMessage(msg)
	// 		break
	// 	}
	// 	msg.Text = "Search for monitored series started"
	// 	b.sendMessage(msg)

	// does not work
	// case "updateAll", "updateall":
	// 	series, err := s.GetSeries(0)
	// 	if err != nil {
	// 		msg.Text = err.Error()
	// 		fmt.Println(err)
	// 		b.sendMessage(msg)
	// 		break
	// 	}
	// 	var allSeriesIDs []int64
	// 	for _, series := range series {
	// 		allSeriesIDs = append(allSeriesIDs, series.ID)
	// 	}
	// 	command := sonarr.CommandRequest{
	// 		Name:      "RefresSeries",
	// 		SeriesIDs: allSeriesIDs,
	// 	}
	// 	_, err = s.SendCommand(&command)
	// 	if err != nil {
	// 		msg.Text = err.Error()
	// 		fmt.Println(err)
	// 		b.sendMessage(msg)
	// 		break
	// 	}
	// 	msg.Text = "Update All started"
	// 	b.sendMessage(msg)

	case "system", "System", "systemstatus", "Systemstatus":
		status, err := s.GetSystemStatus()
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = prettyPrint(status)
		b.sendMessage(msg)

	case "getid", "id":
		msg.Text = fmt.Sprintf("Ваш ID пользователя: %d", chatID)
		b.sendMessage(msg)

	default:
		msg.Text = fmt.Sprintf("Привет, %v!\n", update.Message.From)
		msg.Text += "Доступные команды:\n\n"
		msg.Text += "/q [сериал] - поиск сериала\n"
		msg.Text += "/library [сериал] - управление сериалами\n"
		msg.Text += "/delete [сериал] - удаление сериала\n"
		msg.Text += "/clear - сбросить все отправленные команды\n"
		msg.Text += "/free - свободное место на дисках\n"
		msg.Text += "/up - ближайшие серии на 30 дней вперёд\n"
		msg.Text += "/rss - запустить RSS-синхронизацию\n"
		//msg.Text += "/searchmonitored - searches all monitored series\n"
		//msg.Text += "/updateall - updates metadata and rescans files/folders\n"
		msg.Text += "/system - конфигурация Sonarr\n"
		msg.Text += "/id - ваш Telegram ID"
		b.sendMessage(msg)
	}
}
