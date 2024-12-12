package main

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jedib0t/go-pretty/list"
	"github.com/mattpotok/domek/backend/internal/finance"
)

type TelegramBot struct {
	Bot *bot.Bot
}

func NewTelegramBot(ctx context.Context) (*TelegramBot, error) {
	token := getEnvironmentVariable(TELEGRAM_BOT_TOKEN_ENV)
	b, err := bot.New(token)
	if err != nil {
		return nil, err
	}

	_, err = b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{
				Command:     "/finance_cds",
				Description: "Current CD rates",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/finance_cds", bot.MatchTypeExact, handleFinanceCDs)

	go b.Start(ctx)

	telegramBot := &TelegramBot{
		Bot: b,
	}

	return telegramBot, nil
}

func handleFinanceCDs(ctx context.Context, b *bot.Bot, update *models.Update) {
	institutions, err := finance.FetchCDRates()
	if err != nil {
		// TODO consider sending this as a bot message
		fmt.Println(err)
	}

	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	l.AppendItem("Institutions")
	l.Indent()

	for _, institution := range institutions {
		l.AppendItem(institution.Name)
		l.Indent()

		for _, cd := range institution.CDs {
			line := fmt.Sprintf("%d mo - %s", cd.Term, cd.Rate)
			l.AppendItem(line)
		}
		l.UnIndent()
	}

	text := fmt.Sprintf("<pre>%s</pre>", l.Render())
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
}
