package keyboard

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vladislavprovich/TG-bot/internal/models"
	"github.com/vladislavprovich/TG-bot/pkg"
)

func MainMenu() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create short URL", "create_short_url"),
			tgbotapi.NewInlineKeyboardButtonData("List of all short URLs", "list_short_urls"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Show short URL stats", "show_url_stats"),
			tgbotapi.NewInlineKeyboardButtonData("Settings", "settings"),
		),
	)
	return &keyboard
}

func BackMenu() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_main"),
		),
	)
	return &keyboard
}

func ClearAndBack() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Clear History", "clear_history"),
			tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_main"),
		),
	)
	return &keyboard
}

func CreateURL() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create Random", "rand_url"),

			tgbotapi.NewInlineKeyboardButtonData("Create Custom", "cust_url"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_main"),
		),
	)
	return &keyboard
}

func CreateURLListWithDeleteButtons(urls []*models.GetListResponse) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, url := range urls {
		// only original url. Example: "http://google.com/qwerty" --> "google.com"
		redactionOriginalURL := pkg.OriginalInfo(url.OriginalURL)
		// only short url. Example: "http://host:1111/qwerty" --> "qwerty"
		redactionShortURL := pkg.ShortInfo(url.ShortURL)

		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(redactionOriginalURL, url.OriginalURL),
			// button no usage. TG Api block localhost
			tgbotapi.NewInlineKeyboardButtonData(redactionShortURL, "ignore"),
			tgbotapi.NewInlineKeyboardButtonData("Delete",
				fmt.Sprintf("delete_short_url:%s", url.ShortURL)),
		)
		rows = append(rows, row)
	}

	backRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_main"),
	)
	rows = append(rows, backRow)

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func CreateURLStatusButton(stats []*models.GetURLStatusResponse) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, url := range stats {
		// Only short url. Example: "http://host:1111/qwerty" --> "qwerty".
		redactionShortURL := pkg.ShortInfo(url.ShortURL)
		regi := strconv.Itoa(url.RedirectCount)

		datePart := url.CreatedAt.Format("2006-01-02")
		timePart := url.CreatedAt.Format("15:04")
		formatDate := fmt.Sprintf(datePart + "\n" + timePart)

		row := tgbotapi.NewInlineKeyboardRow(
			// Button no usage. TG Api block localhost.
			tgbotapi.NewInlineKeyboardButtonData(redactionShortURL, "ignore"),
			tgbotapi.NewInlineKeyboardButtonData(formatDate, "ignore"),
			tgbotapi.NewInlineKeyboardButtonData(regi, "ignore"),
		)
		rows = append(rows, row)
	}

	backRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_main"),
	)
	rows = append(rows, backRow)

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}
