package repository

//go:generate mockgen -destination=../../mocks/mock_adminuser.go -package=mocks runclub/internal/domain/repository AdminUserRepository
//go:generate mockgen -destination=../../mocks/mock_adminuserclub.go -package=mocks runclub/internal/domain/repository AdminUserClubRepository
//go:generate mockgen -destination=../../mocks/mock_botstate.go -package=mocks runclub/internal/domain/repository BotStateRepository
//go:generate mockgen -destination=../../mocks/mock_club.go -package=mocks runclub/internal/domain/repository ClubRepository
//go:generate mockgen -destination=../../mocks/mock_customfield.go -package=mocks runclub/internal/domain/repository CustomFieldRepository,CustomFieldValueRepository
//go:generate mockgen -destination=../../mocks/mock_jointrun.go -package=mocks runclub/internal/domain/repository JointRunRepository,JointRunParticipantRepository
//go:generate mockgen -destination=../../mocks/mock_location.go -package=mocks runclub/internal/domain/repository LocationRepository
//go:generate mockgen -destination=../../mocks/mock_member.go -package=mocks runclub/internal/domain/repository MemberRepository,ClubMemberRepository
//go:generate mockgen -destination=../../mocks/mock_notification.go -package=mocks runclub/internal/domain/repository RaceNotificationLogRepository
//go:generate mockgen -destination=../../mocks/mock_race.go -package=mocks runclub/internal/domain/repository RaceRepository,RaceRegistrationRepository
//go:generate mockgen -destination=../../mocks/mock_template.go -package=mocks runclub/internal/domain/repository TemplateRepository
//go:generate mockgen -destination=../../mocks/mock_training.go -package=mocks runclub/internal/domain/repository TrainingRepository,TrainingTrainerRepository,TrainingParticipantRepository
