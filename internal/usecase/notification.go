package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

// MessageSender abstracts sending Telegram messages, implemented by the telegram delivery layer.
type MessageSender interface {
	SendMessage(chatID int64, text string) error
	SendMessageWithKeyboard(chatID int64, text string, keyboard any) error
}

type NotificationUseCase interface {
	SendBirthdayGreetings(ctx context.Context) error
	SendRaceNotifications(ctx context.Context) error
}

type notificationUseCase struct {
	clubRepo       repository.ClubRepository
	memberRepo     repository.MemberRepository
	clubMemberRepo repository.ClubMemberRepository
	raceRepo       repository.RaceRepository
	raceRegRepo    repository.RaceRegistrationRepository
	raceNotifyRepo repository.RaceNotificationLogRepository
	tmplRepo       repository.TemplateRepository
	sender         MessageSender
}

func NewNotificationUseCase(
	clubRepo repository.ClubRepository,
	memberRepo repository.MemberRepository,
	clubMemberRepo repository.ClubMemberRepository,
	raceRepo repository.RaceRepository,
	raceRegRepo repository.RaceRegistrationRepository,
	raceNotifyRepo repository.RaceNotificationLogRepository,
	tmplRepo repository.TemplateRepository,
	sender MessageSender,
) NotificationUseCase {
	return &notificationUseCase{
		clubRepo:       clubRepo,
		memberRepo:     memberRepo,
		clubMemberRepo: clubMemberRepo,
		raceRepo:       raceRepo,
		raceRegRepo:    raceRegRepo,
		raceNotifyRepo: raceNotifyRepo,
		tmplRepo:       tmplRepo,
		sender:         sender,
	}
}

func (uc *notificationUseCase) SendBirthdayGreetings(ctx context.Context) error {
	clubs, err := uc.clubRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list clubs: %w", err)
	}

	now := time.Now()
	month := int(now.Month())
	day := now.Day()

	members, memberErr := uc.memberRepo.ListBirthdayOn(ctx, month, day)
	if memberErr != nil {
		return fmt.Errorf("list birthday members: %w", memberErr)
	}

	if len(members) == 0 {
		return nil
	}

	for _, club := range clubs {
		if !club.BirthdayEnabled {
			continue
		}

		tmpl, tmplErr := uc.tmplRepo.GetByClubAndType(ctx, club.ID, entity.TemplateBirthday)
		if tmplErr != nil {
			continue
		}

		var birthdayNames []string
		for _, m := range members {
			cm, cmErr := uc.clubMemberRepo.GetByClubAndMember(ctx, club.ID, m.ID)
			if cmErr != nil || cm == nil {
				continue
			}
			birthdayNames = append(birthdayNames, m.FIO)
		}

		if len(birthdayNames) == 0 {
			continue
		}

		text := strings.ReplaceAll(tmpl.Content, "{{names}}", strings.Join(birthdayNames, ", "))

		if sendErr := uc.sender.SendMessage(club.TelegramChatID, text); sendErr != nil {
			continue
		}
	}

	return nil
}

func (uc *notificationUseCase) SendRaceNotifications(ctx context.Context) error {
	clubs, err := uc.clubRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list clubs: %w", err)
	}

	now := time.Now()
	from := now.Add(3 * 24 * time.Hour)
	to := now.Add(4 * 24 * time.Hour)
	sentDate := now.Format("2006-01-02")

	for _, club := range clubs {
		if !club.RaceNotifyEnabled {
			continue
		}

		races, racesErr := uc.raceRepo.ListUpcomingByClub(ctx, club.ID, from, to)
		if racesErr != nil {
			continue
		}

		for _, race := range races {
			exists, existsErr := uc.raceNotifyRepo.Exists(ctx, club.ID, race.ID, sentDate)
			if existsErr != nil || exists {
				continue
			}

			tmpl, tmplErr := uc.tmplRepo.GetByClubAndType(ctx, club.ID, entity.TemplateRaceNotify)
			if tmplErr != nil {
				continue
			}

			text := tmpl.Content
			text = strings.ReplaceAll(text, "{{race_name}}", race.Name)
			text = strings.ReplaceAll(text, "{{race_date}}", race.Date.Format("02.01.2006"))
			text = strings.ReplaceAll(text, "{{race_place}}", race.Place)
			text = strings.ReplaceAll(text, "{{race_distances}}", race.Distances)

			if sendErr := uc.sender.SendMessage(club.TelegramChatID, text); sendErr != nil {
				continue
			}

			_ = uc.raceNotifyRepo.Create(ctx, club.ID, race.ID, sentDate)
		}
	}

	return nil
}
