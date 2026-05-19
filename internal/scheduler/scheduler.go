package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"runclub/internal/config"
	"runclub/internal/usecase"
)

const (
	cronTimeout        = 30 * time.Second
	cleanupTimeout     = 60 * time.Second
	defaultCleanupDays = 50
)

type Scheduler struct {
	notifUC    usecase.NotificationUseCase
	trainingUC usecase.TrainingUseCase
	memberUC   usecase.MemberUseCase
	bot        BotMessenger
	logger     *zap.Logger
	cfg        config.Config
	cron       *cron.Cron
}

// BotMessenger defines the interface the scheduler needs to send Telegram messages.
type BotMessenger interface {
	SendConfirmationPrompt(ctx context.Context, trainingID int64) error
}

func NewScheduler(
	notifUC usecase.NotificationUseCase,
	trainingUC usecase.TrainingUseCase,
	memberUC usecase.MemberUseCase,
	bot BotMessenger,
	logger *zap.Logger,
	cfg config.Config,
) *Scheduler {
	return &Scheduler{
		notifUC:    notifUC,
		trainingUC: trainingUC,
		memberUC:   memberUC,
		bot:        bot,
		logger:     logger,
		cfg:        cfg,
	}
}

func (s *Scheduler) Start() {
	s.cron = cron.New(cron.WithSeconds())

	if s.notifUC != nil {
		if _, err := s.cron.AddFunc(s.cfg.BirthdayCron, s.runBirthdayCheck); err != nil {
			s.logger.Error("failed to add birthday cron", zap.Error(err))
		}

		if _, err := s.cron.AddFunc(s.cfg.RaceNotifyCron, s.runRaceNotifications); err != nil {
			s.logger.Error("failed to add race notify cron", zap.Error(err))
		}
	}

	if _, err := s.cron.AddFunc(s.cfg.TrainingConfirmCron, s.runTrainingConfirmation); err != nil {
		s.logger.Error("failed to add training confirm cron", zap.Error(err))
	}

	if _, err := s.cron.AddFunc(s.cfg.MemberCleanupCron, s.runMemberCleanup); err != nil {
		s.logger.Error("failed to add member cleanup cron", zap.Error(err))
	}

	s.cron.Start()
	s.logger.Info("scheduler started")
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		s.logger.Info("scheduler stopped")
	}
}

func (s *Scheduler) runBirthdayCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), cronTimeout)
	defer cancel()

	if err := s.notifUC.SendBirthdayGreetings(ctx); err != nil {
		s.logger.Error("birthday check failed", zap.Error(err))
	}
}

func (s *Scheduler) runRaceNotifications() {
	ctx, cancel := context.WithTimeout(context.Background(), cronTimeout)
	defer cancel()

	if err := s.notifUC.SendRaceNotifications(ctx); err != nil {
		s.logger.Error("race notification failed", zap.Error(err))
	}
}

func (s *Scheduler) runTrainingConfirmation() {
	ctx, cancel := context.WithTimeout(context.Background(), cronTimeout)
	defer cancel()

	trainings, err := s.trainingUC.FindTrainingsNeedingConfirmation(ctx)
	if err != nil {
		s.logger.Error("find trainings needing confirmation failed", zap.Error(err))
		return
	}

	for _, t := range trainings {
		if err = s.bot.SendConfirmationPrompt(ctx, t.ID); err != nil {
			s.logger.Error("send confirmation prompt failed", zap.Int64("training_id", t.ID), zap.Error(err))
		}
	}
}

func (s *Scheduler) runMemberCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	days := s.cfg.MemberCleanupDays
	if days <= 0 {
		days = defaultCleanupDays
	}

	deleted, err := s.memberUC.CleanupOrphanMembers(ctx, days)
	if err != nil {
		s.logger.Error("member cleanup failed", zap.Error(err))
		return
	}
	if deleted > 0 {
		s.logger.Info("cleaned up orphan members", zap.Int("deleted", deleted), zap.Int("older_than_days", days))
	}
}
