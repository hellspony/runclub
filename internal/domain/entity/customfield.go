package entity

type CustomField struct {
	ID        int64  `json:"id"         db:"id"`
	ClubID    int64  `json:"club_id"    db:"club_id"`
	Name      string `json:"name"       db:"name"`
	Required  bool   `json:"required"   db:"required"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
}

type CustomFieldValue struct {
	ID            int64  `json:"id"              db:"id"`
	MemberID      int64  `json:"member_id"       db:"member_id"`
	CustomFieldID int64  `json:"custom_field_id" db:"custom_field_id"`
	Value         string `json:"value"           db:"value"`
}
