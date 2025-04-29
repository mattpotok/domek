package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mattpotok/domek/backend/internal/common"
	"github.com/mattpotok/domek/backend/internal/database"
)

const telegram_bot_token_env = "TELEGRAM_BOT_TOKEN"

type Telegram struct {
	bot           *bot.Bot
	db            *database.DB
	state         int16
	stateHandlers map[int16]bot.HandlerFunc

	// Garden forms
	// TODO consider encapsulating all of these within a garden specific struct
	vegetable      database.Vegetable
	vegetables     []database.Vegetable
	vegetableField string
}

func NewTelegram(ctx context.Context, cfg *common.TelegramConfig, db *database.DB) (*Telegram, error) {
	telegram := &Telegram{
		db:            db,
		state:         stateDefault,
		stateHandlers: make(map[int16]bot.HandlerFunc),
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(telegram.handleMessage),
	}

	var err error
	telegram.bot, err = bot.New(cfg.BotToken, opts...)
	if err != nil {
		return nil, err
	}

	commands := []models.BotCommand{}
	commands = append(commands, telegram.registerGardenHandlers()...)
	commands = append(commands, telegram.registerFinanceHandlers()...)
	_, err = telegram.bot.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: commands})
	if err != nil {
		return nil, err
	}

	go telegram.bot.Start(ctx)

	return telegram, nil
}

func (telegram *Telegram) handleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling user message...")

	if handler, ok := telegram.stateHandlers[telegram.state]; ok {
		handler(ctx, b, update)
	} else {
		group := (telegram.state & stateMaskGroup) >> 8
		command := (telegram.state & stateMaskCommand) >> 4
		step := (telegram.state & stateMaskStep)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("State 0b%04b_%04b_%04b does not have a handler", group, command, step),
		})

		telegram.state = stateDefault
	}
}
