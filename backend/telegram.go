package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jedib0t/go-pretty/list"
	"github.com/mattpotok/domek/backend/internal/database"
	"github.com/mattpotok/domek/backend/internal/finance"
)

type state string

// TODO consider using a bitwise string to allow for subforms
// Example 0x0001 0001
const (
	stateDefault            state = "default"
	stateGardenAddVegetable       = "gardenAddVegetable"
	stateGardenAddVariety         = "gardenAddVariety"
	stateGardenAddHasFruit        = "gardenAddHasFruit"
)

type formGardenAdd struct {
	Variety   string
	Vegetable string
	HasFruit  bool
	Year      int
}

type TelegramBot struct {
	bot   *bot.Bot
	db    *database.DB
	state state

	// Forms
	vegetable database.Vegetable
}

func NewTelegramBot(ctx context.Context, db *database.DB) (*TelegramBot, error) {
	tgb := &TelegramBot{state: stateDefault, db: db}

	token := getEnvironmentVariable(TELEGRAM_BOT_TOKEN_ENV)

	opts := []bot.Option{
		bot.WithDefaultHandler(tgb.handleDefault),
	}

	var err error
	tgb.bot, err = bot.New(token, opts...)
	if err != nil {
		return nil, err
	}

	_, err = tgb.bot.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{
				Command:     "/garden_add",
				Description: "Add a plant to the garden",
			},
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
		},
	})
	if err != nil {
		return nil, err
	}

	// Finance commands
	tgb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_cds", bot.MatchTypeExact, handleFinanceCDs)
	tgb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_effective_rates", bot.MatchTypePrefix, handleFinanceEffectiveRates)
	tgb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/finance_savings", bot.MatchTypeExact, handleFinanceSavings)

	// Garden commands
	tgb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/garden_add", bot.MatchTypePrefix, tgb.handleGardenAdd)
	// tgb.bot.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regexGardenAdd, tgb.handleGardenAdd)

	go tgb.bot.Start(ctx)

	return tgb, nil
}

func (tgb *TelegramBot) handleDefault(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling default...")

	switch tgb.state {
	// TODO consider
	case stateGardenAddVegetable:
		tgb.handleGardenAddName(ctx, b, update)
	case stateGardenAddVariety:
		tgb.handleGardenAddVariety(ctx, b, update)
	case stateGardenAddHasFruit:
		tgb.handleGardenAddHasFruit(ctx, b, update)

	default:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Unknown command",
		})
	}
}

func (tgb *TelegramBot) handleGardenAdd(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling garden add...")

	// TODO make this a constant
	// TODO figure if the 'P' is necessary
	// re := regexp.MustCompile(`^/garden_add(\s+(?P<year>\d{4}))?$`)
	re := regexp.MustCompile(`^/garden_add(\s+(?<year>\d{4}))?$`)

	matches := re.FindStringSubmatch(update.Message.Text)
	if len(matches) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Incorrect format. Please use <pre>/garden_add YYYY</pre>",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	idx := re.SubexpIndex("year")
	if idx == -1 {
		tgb.vegetable.Year = time.Now().Year()
	} else {
		tgb.vegetable.Year, _ = strconv.Atoi(matches[idx])
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Enter vegetable",
	})

	tgb.state = stateGardenAddVegetable
}

func (tgb *TelegramBot) handleGardenAddName(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling garden add name...")

	// TODO make this a constant
	re := regexp.MustCompile(`^(?<name>[A-Za-z]+(?: [A-Za-z]+)?)$`)

	matches := re.FindStringSubmatch(update.Message.Text)
	if len(matches) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Invalid name. Please use space separated words, example 'a B'.",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	name := matches[re.SubexpIndex("name")]
	tgb.vegetable.Name = strings.ToLower(name)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Enter %s variety", tgb.vegetable.Name),
	})

	tgb.state = stateGardenAddVariety
}

func (tgb *TelegramBot) handleGardenAddVariety(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling garden add variety...")

	re := regexp.MustCompile(`^(?<variety>(?:[A-Za-z]+(?: [A-Za-z]+)?)|-)$`)

	matches := re.FindStringSubmatch(update.Message.Text)
	if len(matches) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Invalid variety. Please use space separated words, example 'A B' or '-' for no variety.",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	variety := matches[re.SubexpIndex("variety")]
	if variety == "-" {
		variety = ""
	}
	tgb.vegetable.Variety = strings.ToLower(variety)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Enter whether %s has fruit (yes, no)", tgb.vegetable.Name),
	})

	tgb.state = stateGardenAddHasFruit
}

// TODO test this works and then move to it's own file
func (tgb *TelegramBot) handleGardenAddHasFruit(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling garden add hasFruit...")

	re := regexp.MustCompile(`^(?<hasFruit>yes|no)$`)

	matches := re.FindStringSubmatch(update.Message.Text)
	if len(matches) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Invalid input. Must be either 'yes' or 'no'.",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	hasFruit := matches[re.SubexpIndex("hasFruit")]
	if hasFruit == "yes" {
		tgb.vegetable.HasFruit = 1
	} else {
		tgb.vegetable.HasFruit = 0
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Vegetable from %+v", tgb.vegetable),
	})

	vegetable := database.Vegetable{
		Id:       0,
		Name:     tgb.vegetable.Name,
		HasFruit: tgb.vegetable.HasFruit,
		Quantity: 0,
		Variety:  tgb.vegetable.Variety,
		Year:     tgb.vegetable.Year,
		Yield:    0,
	}
	err := tgb.db.InsertVegetable(&vegetable)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Unable to add vegetable to database - %s", err.Error()),
		})
	}

	tgb.state = stateDefault
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

func handleFinanceEffectiveRates(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling effective rates...")

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
