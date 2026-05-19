package tgutil

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"runclub/internal/domain/entity"
)

func ClubKeyboard(clubs []entity.ClubMember, prefix string) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(clubs))
	for _, c := range clubs {
		data := Encode(prefix, "club", strconv.FormatInt(c.ClubID, 10))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(string(c.Role)+" — club #"+strconv.FormatInt(c.ClubID, 10), data),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func ClubKeyboardFromClubs(clubs []entity.Club, prefix string) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(clubs))
	for _, c := range clubs {
		data := Encode(prefix, "club", strconv.FormatInt(c.ID, 10))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(c.Name, data),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func LocationKeyboard(locations []entity.Location, clubID int64) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(locations)+1)
	for _, loc := range locations {
		data := Encode("sel", "loc", strconv.FormatInt(loc.ID, 10))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(loc.Name, data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("New location", Encode("new", "loc", strconv.FormatInt(clubID, 10))),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrainerKeyboard(trainers []entity.Member, selected []int64, _ int64) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(trainers)+1)

	selectedMap := make(map[int64]bool, len(selected))
	for _, id := range selected {
		selectedMap[id] = true
	}

	for _, t := range trainers {
		prefix := "addsel"
		if selectedMap[t.ID] {
			prefix = "remsel"
		}
		check := "  "
		if selectedMap[t.ID] {
			check = "✓ "
		}
		data := Encode(prefix, "trainer", strconv.FormatInt(t.ID, 10))
		label := fmt.Sprintf("%s%s", check, DisplayName(t))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, data),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Done", Encode("done", "trainer")),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func ParticipationKeyboard(trainingID int64, isJoined bool) tgbotapi.InlineKeyboardMarkup {
	var btn tgbotapi.InlineKeyboardButton
	if isJoined {
		btn = tgbotapi.NewInlineKeyboardButtonData(
			"Не иду",
			Encode("leave", "training", strconv.FormatInt(trainingID, 10)),
		)
	} else {
		btn = tgbotapi.NewInlineKeyboardButtonData("Иду", Encode("join", "training", strconv.FormatInt(trainingID, 10)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
}

func ParticipationKeyboardJointRun(runID int64, isJoined bool) tgbotapi.InlineKeyboardMarkup {
	var btn tgbotapi.InlineKeyboardButton
	if isJoined {
		btn = tgbotapi.NewInlineKeyboardButtonData("Не иду", Encode("leave", "jointrun", strconv.FormatInt(runID, 10)))
	} else {
		btn = tgbotapi.NewInlineKeyboardButtonData("Иду", Encode("join", "jointrun", strconv.FormatInt(runID, 10)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
}

func ConfirmationKeyboard(trainingID int64) tgbotapi.InlineKeyboardMarkup {
	idStr := strconv.FormatInt(trainingID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Готово", Encode("cfm", "training", idStr)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить фото", Encode("photo", "training", idStr)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("+ Участник", Encode("addsel", "training", idStr)),
			tgbotapi.NewInlineKeyboardButtonData("- Участник", Encode("remsel", "training", idStr)),
		),
	)
}

func WelcomeKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", Encode("wel", "yes")),
			tgbotapi.NewInlineKeyboardButtonData("Нет", Encode("wel", "no")),
		),
	)
}

func MemberSelectKeyboard(members []entity.Member, prefix string, entityID int64) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(members))
	idStr := strconv.FormatInt(entityID, 10)
	for _, m := range members {
		data := Encode(prefix, "training", idStr, strconv.FormatInt(m.ID, 10))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(DisplayName(m), data),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func DisplayName(m entity.Member) string {
	if m.FIO != "" {
		return m.FIO
	}
	if m.TelegramUsername != "" {
		return "@" + m.TelegramUsername
	}
	return fmt.Sprintf("user#%d", m.ID)
}
