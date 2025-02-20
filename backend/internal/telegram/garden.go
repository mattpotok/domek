package telegram

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jedib0t/go-pretty/list"
	"github.com/mattpotok/domek/backend/internal/database"
)

const (
	stateGardenAddHarvest int16 = stateGarden + iota<<4
	stateGardenAddVegetable
	stateGardenUpdateVegetable
)

const (
	stateGardenAddHarvestSelection int16 = stateGardenAddHarvest + iota
	stateGardenAddHarvestAmount
)

const (
	stateGardenAddVegetableName int16 = stateGardenAddVegetable + iota
	stateGardenAddVegetableVariety
	stateGardenAddVegetableQuantity
	stateGardenAddVegetableHasFruit
)

const (
	stateGardenUpdateVegetableSelection int16 = stateGardenUpdateVegetable + iota
	stateGardenUpdateVegetableField
	stateGardenUpdateVegetableValue
)

var vegetableUpdateableFields = []string{"quantity"}

var regexGardenDescribeVegetables = regexp.MustCompile(`^/garden_describe_vegetables(?:\s+(?<year>\d{4}))?$`)
var regexVegetableHasFruit = regexp.MustCompile(`^(?<hasFruit>(?i)yes|no)$`)
var regexVegetableName = regexp.MustCompile(`^(?<name>[A-Za-z]+(?: [A-Za-z]+)?)$`)
var regexVegetableVariety = regexp.MustCompile(`^(?<variety>[A-Za-z]+(?: [A-Za-z]+)?)$`)

func (telegram *Telegram) registerGardenHandlers() []models.BotCommand {
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/garden_add_harvest", bot.MatchTypeExact, telegram.handleGardenAddHarvest)
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/garden_add_vegetable", bot.MatchTypeExact, telegram.handleGardenAddVegetable)
	// telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/garden_describe_vegetables", bot.MatchTypeExact, telegram.handleGardenDescribeVegetables)
	telegram.bot.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regexGardenDescribeVegetables, telegram.handleGardenDescribeVegetables)
	telegram.bot.RegisterHandler(bot.HandlerTypeMessageText, "/garden_update_vegetable", bot.MatchTypeExact, telegram.handleGardenUpdateVegetable)

	// State handlers for '/garden_add_harvest'
	telegram.stateHandlers[stateGardenAddHarvestSelection] = telegram.handleGardenAddHarvestSelection
	telegram.stateHandlers[stateGardenAddHarvestAmount] = telegram.handleGardenAddHarvestAmount

	// State handlers for '/garden_add_vegetable'
	telegram.stateHandlers[stateGardenAddVegetableName] = telegram.handleGardenAddVegetableName
	telegram.stateHandlers[stateGardenAddVegetableVariety] = telegram.handleGardenAddVegetableVariety
	telegram.stateHandlers[stateGardenAddVegetableQuantity] = telegram.handleGardenAddVegetableQuantity
	telegram.stateHandlers[stateGardenAddVegetableHasFruit] = telegram.handleGardenAddVegetableHasFruit

	// State handlers for '/garden_update_vegetable'
	telegram.stateHandlers[stateGardenUpdateVegetableSelection] = telegram.handleGardenUpdateVegetableSelection
	telegram.stateHandlers[stateGardenUpdateVegetableField] = telegram.handleGardenUpdateVegetableField
	telegram.stateHandlers[stateGardenUpdateVegetableValue] = telegram.handleGardenUpdateVegetableValue

	return []models.BotCommand{
		{
			Command:     "/garden_add_harvest",
			Description: "Record harvest for vegetable",
		},
		{
			Command:     "/garden_add_vegetable",
			Description: "Add a vegetable to garden",
		},
		{
			Command:     "/garden_describe_vegetables",
			Description: "Describe all vegetables for a season",
		},
		{
			Command:     "/garden_update_vegetable",
			Description: "Update a vegetable",
		},
	}
}

func (telegram *Telegram) handleGardenAddHarvest(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_harvest'...")

	year := time.Now().Year()

	var err error
	telegram.vegetables, err = telegram.db.GetVegetablesByYear(year)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Unable to query database for vegetables",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	sortVegetablesByNameAndVariety(telegram.vegetables)

	list := listVegetables(telegram.vegetables)
	text := fmt.Sprintf("Select a vegetable\n<pre>%s</pre>", list)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	telegram.state = stateGardenAddHarvestSelection
}

func (telegram *Telegram) handleGardenAddHarvestSelection(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_harvest' step 'select' ...")

	selection, err := strconv.ParseUint(update.Message.Text, 10, 0)
	if err != nil || selection >= uint64(len(telegram.vegetables)) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Invalid selection. Please use a number between 0 and %d.", len(telegram.vegetables)-1),
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	telegram.vegetable = telegram.vegetables[selection]
	description := describeVegetables(telegram.vegetable)
	text := fmt.Sprintf("%s\nEnter amount harvested", description)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	telegram.state = stateGardenAddHarvestAmount
}

func (telegram *Telegram) handleGardenAddHarvestAmount(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_harvest' step 'amount' ...")

	amount, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Invalid amount. Please use a number",
		})
	}

	telegram.vegetable.Yield += amount
	if err := telegram.db.UpdateVegetable(telegram.vegetable); err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Unable to add harvest",
		})
	}

	text := describeVegetables(telegram.vegetable)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	telegram.state = stateDefault
}

func (telegram *Telegram) handleGardenAddVegetable(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_vegetable'...")

	telegram.vegetable.Year = time.Now().Year()

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Enter vegetable",
	})

	telegram.state = stateGardenAddVegetableName
}

func (telegram *Telegram) handleGardenAddVegetableName(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_vegetable' step 'name'...")

	matches := regexVegetableName.FindStringSubmatch(update.Message.Text)
	if len(matches) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Invalid name. Please use space separated words.",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	name := matches[regexVegetableName.SubexpIndex("name")]
	telegram.vegetable.Name = strings.ToLower(name)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Enter %s variety", telegram.vegetable.Name),
	})

	telegram.state = stateGardenAddVegetableVariety
}

func (telegram *Telegram) handleGardenAddVegetableVariety(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_vegetable' step 'variety'...")

	matches := regexVegetableVariety.FindStringSubmatch(update.Message.Text)
	if len(matches) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Invalid variety. Please use space separated words, example 'A B' or '-' for no variety.",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	variety := matches[regexVegetableVariety.SubexpIndex("variety")]
	telegram.vegetable.Variety = strings.ToLower(variety)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Enter quantity of %s planted", telegram.vegetable.Name),
	})

	telegram.state = stateGardenAddVegetableQuantity
}

func (telegram *Telegram) handleGardenAddVegetableQuantity(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_vegetable' step 'hasFruit'...")

	quantity, err := strconv.ParseUint(update.Message.Text, 10, 0)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Invalid input. Must be a non-negative number.",
			ParseMode: models.ParseModeHTML,
		})
	}

	telegram.vegetable.Quantity = int(quantity)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Enter whether %s has fruit (yes, no)", telegram.vegetable.Name),
	})

	telegram.state = stateGardenAddVegetableHasFruit
}

func (telegram *Telegram) handleGardenAddVegetableHasFruit(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_add_vegetable' step 'hasFruit'...")

	matches := regexVegetableHasFruit.FindStringSubmatch(update.Message.Text)
	if len(matches) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Invalid input. Must be either 'yes' or 'no'.",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	hasFruit := matches[regexVegetableHasFruit.SubexpIndex("hasFruit")]
	hasFruit = strings.ToLower(hasFruit)
	telegram.vegetable.HasFruit = hasFruit == "yes"

	text := "Added vegetable\n" + describeVegetables(telegram.vegetable)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	vegetable := database.Vegetable{
		Id:       0,
		Name:     telegram.vegetable.Name,
		HasFruit: telegram.vegetable.HasFruit,
		Quantity: telegram.vegetable.Quantity,
		Variety:  telegram.vegetable.Variety,
		Year:     telegram.vegetable.Year,
		Yield:    0,
	}
	err := telegram.db.InsertVegetable(&vegetable)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Unable to add vegetable to database - %s", err.Error()),
		})
	}

	telegram.state = stateDefault
}

func (telegram *Telegram) handleGardenDescribeVegetables(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_describe_vegetables'...")

	year := time.Now().Year()

	matches := regexGardenDescribeVegetables.FindStringSubmatch(update.Message.Text)
	idx := regexGardenDescribeVegetables.SubexpIndex("year")
	if matches[idx] != "" {
		year, _ = strconv.Atoi(matches[idx])
	}

	vegetables, err := telegram.db.GetVegetablesByYear(year)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Unable to query database for vegetables",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	sortVegetablesByNameAndVariety(vegetables)

	text := describeVegetables(vegetables...)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
}

func (telegram *Telegram) handleGardenUpdateVegetable(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_update_vegetable'...")

	year := time.Now().Year()

	var err error
	telegram.vegetables, err = telegram.db.GetVegetablesByYear(year)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Unable to query database for vegetables",
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	sortVegetablesByNameAndVariety(telegram.vegetables)

	list := listVegetables(telegram.vegetables)
	text := fmt.Sprintf("Select a vegetable\n<pre>%s</pre>", list)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	telegram.state = stateGardenUpdateVegetableSelection
}

func (telegram *Telegram) handleGardenUpdateVegetableSelection(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_update_vegetable' step 'selection' ...")

	selection, err := strconv.ParseUint(update.Message.Text, 10, 0)
	if err != nil || selection >= uint64(len(telegram.vegetables)) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Invalid selection. Please use a number between 0 and %d.", len(telegram.vegetables)-1),
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	list := ""
	for i, field := range vegetableUpdateableFields {
		list += fmt.Sprintf("%d. %s\n", i, field)
	}

	telegram.vegetable = telegram.vegetables[selection]
	text := fmt.Sprintf("Select a field to update\n<pre>%s</pre>", list)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	telegram.state = stateGardenUpdateVegetableField
}

func (telegram *Telegram) handleGardenUpdateVegetableField(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_update_vegetable' step 'field' ...")

	selection, err := strconv.ParseUint(update.Message.Text, 10, 0)
	if err != nil || selection >= uint64(len(vegetableUpdateableFields)) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Invalid selection. Please use a number between 0 and %d.", len(vegetableUpdateableFields)-1),
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	telegram.vegetableField = vegetableUpdateableFields[selection]
	description := describeVegetables(telegram.vegetable)
	text := fmt.Sprintf("%s\nEnter new value for '%s'", description, telegram.vegetableField)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	telegram.state = stateGardenUpdateVegetableValue
}

func (telegram *Telegram) handleGardenUpdateVegetableValue(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("Handling command '/garden_update_vegetable' step 'value' ...")

	// TODO support other fields here
	switch telegram.vegetableField {
	case "quantity":
		quantity, err := strconv.ParseUint(update.Message.Text, 10, 0)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				Text:      "Invalid input. Must be a non-negative number.",
				ParseMode: models.ParseModeHTML,
			})
		}

		telegram.vegetable.Quantity = int(quantity)
	}

	if err := telegram.db.UpdateVegetable(telegram.vegetable); err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Unable to update vegetable",
		})
	}

	text := describeVegetables(telegram.vegetable)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	telegram.state = stateDefault
}

// TODO consider making this configurable to allow for any field selection
func describeVegetables(vegetables ...database.Vegetable) string {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	for _, veg := range vegetables {
		name := veg.Name
		if len(veg.Variety) > 0 {
			name += " (" + veg.Variety + ")"
		}

		l.AppendItem(name)
		l.Indent()
		l.AppendItem(fmt.Sprintf("hasFruit - %t", veg.HasFruit))
		l.AppendItem(fmt.Sprintf("quantity - %d", veg.Quantity))
		l.AppendItem(fmt.Sprintf("total yield - %d", veg.Yield))

		avgYield := 0.0
		if veg.Quantity > 0 {
			avgYield = float64(veg.Yield) / float64(veg.Quantity)
		}
		l.AppendItem(fmt.Sprintf("avg yield - %0.2f", avgYield))

		l.UnIndent()
	}

	return fmt.Sprintf("<pre>%s</pre>", l.Render())
}

func listVegetables(vegetables []database.Vegetable) string {
	list := ""
	for i, veg := range vegetables {
		name := veg.Name
		if len(veg.Variety) > 0 {
			name += " (" + veg.Variety + ")"
		}

		list += fmt.Sprintf("%d. %s\n", i, name)
	}

	return list
}

func sortVegetablesByNameAndVariety(vegetables []database.Vegetable) {
	slices.SortFunc(vegetables, func(a database.Vegetable, b database.Vegetable) int {
		if a.Name != b.Name {
			return strings.Compare(a.Name, b.Name)
		}

		return strings.Compare(a.Variety, b.Variety)
	})
}
