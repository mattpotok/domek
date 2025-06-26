package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jedib0t/go-pretty/list"
	"github.com/mattpotok/domek/backend/internal/finance"
)

func (telegram *Telegram) registerFinanceHandlers() []models.BotCommand {
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_cds", bot.MatchTypeExact, telegram.handleFinanceCDs)
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_effective_rates", bot.MatchTypePrefix, handleFinanceEffectiveRates)
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_savings", bot.MatchTypeExact, telegram.handleFinanceSavings)

	return []models.BotCommand{
		{
			Command:     "/finance_cds",
			Description: "Current CD rates",
		},
		{
			Command:     "/finance_savings",
			Description: "Current savings rates",
		},
		{
			Command:     "/finance_effective_rates",
			Description: "Compute effective rates after taxes",
		},
	}
}

func (telegram *Telegram) handleFinanceCDs(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/finance_cds_v2'...")

	accounts := finance.GetCdAccounts(telegram.client)

	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	l.AppendItem("Institutions")
	l.Indent()

	for _, result := range accounts {
		l.Indent()
		if account, err := result.Match(); err != nil {
			line := fmt.Sprintf("%s - %s", account.Name, err)
			l.AppendItem(line)
		} else {
			l.AppendItem(account.Name)
			l.Indent()

			for _, cd := range account.CDs {
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

func handleFinanceEffectiveRates(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/finance_effective rates'...")

	var rate float64 = 0.0
	_, err := fmt.Sscanf(update.Message.Text, "/finance_effective_rates %f", &rate)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			ParseMode: models.ParseModeHTML,
			Text:      "Incorrect format. Please use <pre>/finance_effective_rates RATE</pre>",
		})
		return
	}

	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	l.AppendItem(fmt.Sprintf("Effective rates for %.2f%%", rate))
	l.Indent()

	for _, taxBracket := range finance.TaxBrackets {
		effectiveRate := rate * (1 - taxBracket/100.0)
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

func (telegram *Telegram) handleFinanceSavings(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/finance_savings'...")

	accounts := finance.GetSavingAccounts(telegram.client)

	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	l.AppendItem("Institutions")
	l.Indent()

	for _, result := range accounts {
		l.Indent()

		if account, err := result.Match(); err != nil {
			line := fmt.Sprintf("%s - %s", account.Name, err)
			l.AppendItem(line)
		} else {
			line := fmt.Sprintf("%s - %s%%", account.Name, account.Rate)
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
