package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

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
			{
				Command:     "/finance_savings",
				Description: "Current savings rates",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/finance_cds", bot.MatchTypeExact, handleFinanceCDs)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/finance_savings", bot.MatchTypeExact, handleFinanceSavings)

	go b.Start(ctx)

	telegramBot := &TelegramBot{
		Bot: b,
	}

	return telegramBot, nil
}

func handleFinanceCDs(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling cds...")

	institutions, err := finance.FetchCDRates()
	if err != nil {
		// TODO improve the error message
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
	}

	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	l.AppendItem("Institutions")
	l.Indent()

	for _, result := range institutions {
		l.Indent()

		if institution, err := result.Match(); err != nil {
			line := fmt.Sprintf("%s - %s", institution.Name, err)
			l.AppendItem(line)
		} else {
			l.AppendItem(institution.Name)

			l.Indent()
			for _, cd := range institution.CDs {
				rate := cd.Rate + "%"
				if cd.Rate == "" {
					rate = "---"
				}
				line := fmt.Sprintf("%d mo - %s", cd.Term, rate)
				l.AppendItem(line)
			}
			l.UnIndent()
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

func handleFinanceSavings(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling savings...")

	institutions, err := finance.FetchSavingsAccounts()
	if err != nil {
		// TODO improve the error message
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
	}

	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	l.AppendItem("Institutions")
	l.Indent()

	var maxRate float64 = 0.0
	for _, result := range institutions {
		l.Indent()

		if institution, err := result.Match(); err != nil {
			line := fmt.Sprintf("%s - %s", institution.Name, err)
			l.AppendItem(line)
		} else {
			rate, err := strconv.ParseFloat(institution.Savings, 64)
			if err == nil && rate > maxRate {
				maxRate = rate
			}

			line := fmt.Sprintf("%s - %s%%", institution.Name, institution.Savings)
			l.AppendItem(line)
		}

		l.UnIndent()
	}

	l.UnIndent()
	l.AppendItem(fmt.Sprintf("Effective rates for %.2f%%", maxRate))
	l.Indent()

	for _, taxBracket := range finance.TaxBrackets {
		effectiveRate := maxRate * (1 - taxBracket/100.0)
		line := fmt.Sprintf("At %.1f%% - %.2f%%", taxBracket, effectiveRate)
		l.AppendItem(line)
	}

	text := fmt.Sprintf("<pre>%s</pre>", l.Render())
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
}
