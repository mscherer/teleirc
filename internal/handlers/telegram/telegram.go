// Package telegram handles all Telegram-side logic.
package telegram

import (

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ritlug/teleirc/internal"
)

/*
Client contains information for the Telegram bridge, including
the TelegramSettings needed to run the bot
*/
type Client struct {
	api       *tgbotapi.BotAPI
	Settings  internal.TelegramSettings
	logger    internal.DebugLogger
	sendToIrc func(string)
}

/*
NewClient creates a new Telegram bot client
*/
func NewClient(settings internal.TelegramSettings, tgapi *tgbotapi.BotAPI, debug internal.DebugLogger) *Client {
	debug.LogInfo("Creating new Telegram bot client...")
	return &Client{api: tgapi, Settings: settings, logger: debug}
}

/*
SendMessage sends a message to the Telegram channel specified in the settings
*/
func (tg *Client) SendMessage(msg string) {
	newMsg := tgbotapi.NewMessage(tg.Settings.ChatID, "")
	newMsg.Text = msg

	if _, err := tg.api.Send(newMsg); err != nil {
		var attempts int = 0
		// Try resending 3 times if the message is successfully sent
		for err != nil && attempts < 3 {
			tg.logger.LogError(err)
			attempts++
			_, err = tg.api.Send(newMsg)
		}
	}
}

/*
StartBot adds necessary handlers to the client and then connects,
returning any errors that occur
*/
func (tg *Client) StartBot(errChan chan<- error, sendMessage func(string)) {
	tg.logger.LogInfo("Starting up Telegram bot...")
	var err error
	tg.api, err = tgbotapi.NewBotAPI(tg.Settings.Token)
	if err != nil {
		tg.logger.LogError(err)
		errChan <- err
	}
	tg.sendToIrc = sendMessage

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tg.api.GetUpdatesChan(u)
	if err != nil {
		errChan <- err
		tg.logger.LogError(err)
	}

	// TODO: Move these lines into the updateHandler when available
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Document != nil {
			documentHandler(tg, update.Message)
		} else {
			messageHandler(tg, update)
		}

	}
	errChan <- nil
}