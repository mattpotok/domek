package telegram

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

func (telegram *Telegram) registerFinanceHandlers() []models.BotCommand {
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_cds", bot.MatchTypeExact, handleFinanceCDs)
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_effective_rates", bot.MatchTypePrefix, handleFinanceEffectiveRates)
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_savings", bot.MatchTypeExact, handleFinanceSavings)

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

func handleFinanceCDs(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling comand '/finance_cds'...")

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

func handleFinanceSavings(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/finance_savings'...")

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
