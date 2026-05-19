package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPPort            int    `yaml:"http_port"              env:"HTTP_PORT"              envDefault:"8080"`
	StaticDir           string `yaml:"static_dir"             env:"STATIC_DIR"             envDefault:"./web/admin/dist"`
	DBPath              string `yaml:"db_path"                env:"DB_PATH"                envDefault:"./data/runclub.db"`
	TelegramToken       string `yaml:"telegram_token"         env:"TELEGRAM_TOKEN"`
	WebhookURL          string `yaml:"webhook_url"            env:"WEBHOOK_URL"`
	AdminUser           string `yaml:"admin_user"             env:"ADMIN_USER"             envDefault:"admin"`
	AdminPass           string `yaml:"admin_pass"             env:"ADMIN_PASS:-changeme"   envDefault:"changeme"`
	AdminRole           string `yaml:"admin_role"             env:"ADMIN_ROLE"             envDefault:"superadmin"`
	JWTSecret           string `yaml:"jwt_secret"             env:"JWT_SECRET:-secret"     envDefault:"secret"`
	BirthdayCron        string `yaml:"birthday_cron"          env:"BIRTHDAY_CRON"          envDefault:"0 0 8 * * *"`
	RaceNotifyCron      string `yaml:"race_notify_cron"       env:"RACE_NOTIFY_CRON"       envDefault:"0 0 10 * * 4"`
	TrainingConfirmCron string `yaml:"training_confirm_cron"  env:"TRAINING_CONFIRM_CRON"  envDefault:"0 */5 * * * *"`
	MemberCleanupCron   string `yaml:"member_cleanup_cron"    env:"MEMBER_CLEANUP_CRON"    envDefault:"0 0 3 * * *"`
	MemberCleanupDays   int    `yaml:"member_cleanup_days"    env:"MEMBER_CLEANUP_DAYS"    envDefault:"50"`
	LogLevel            string `yaml:"log_level"              env:"LOG_LEVEL"              envDefault:"info"`
	InitClubName        string `yaml:"init_club_name"         env:"INIT_CLUB_NAME"`
	InitClubChatID      int64  `yaml:"init_club_chat_id"      env:"INIT_CLUB_CHAT_ID"      envDefault:"0"`
	InitAdminTelegramID int64  `yaml:"init_admin_telegram_id" env:"INIT_ADMIN_TELEGRAM_ID" envDefault:"0"`
	InitAdminUsername   string `yaml:"init_admin_username"    env:"INIT_ADMIN_USERNAME"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Read from config.yaml if it exists.
	data, err := os.ReadFile("config.yaml")
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config.yaml: %w", err)
	}
	if err == nil {
		if err = yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config.yaml: %w", err)
		}
	}

	// Override with environment variables.
	if err = env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	return cfg, nil
}
