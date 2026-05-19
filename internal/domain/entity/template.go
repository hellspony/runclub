package entity

type TemplateType string

const (
	TemplateWelcome      TemplateType = "welcome"
	TemplateBirthday     TemplateType = "birthday"
	TemplateRaceNotify   TemplateType = "race_notify"
	TemplateTrainingNew  TemplateType = "training_new"
	TemplateTrainingDone TemplateType = "training_done"
	TemplateJointRunNew  TemplateType = "jointrun_new"
)

type Template struct {
	ID      int64        `json:"id"      db:"id"`
	ClubID  int64        `json:"club_id" db:"club_id"`
	Type    TemplateType `json:"type"    db:"type"`
	Name    string       `json:"name"    db:"name"`
	Content string       `json:"content" db:"content"`
}
