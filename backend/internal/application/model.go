// Package application — заявки пользователей на соревнования и их рассмотрение.
package application

import "time"

type Status string

const (
	StatusPending   Status = "pending"   // на рассмотрении
	StatusApproved  Status = "approved"  // одобрена
	StatusRejected  Status = "rejected"  // отклонена
	StatusCancelled Status = "cancelled" // отменена пользователем
)

// Application — заявка с данными о соревновании и пользователе для списков.
type Application struct {
	ID               int64     `json:"id"`
	CompetitionID    int64     `json:"competition_id"`
	CompetitionTitle string    `json:"competition_title"`
	StartsOn         string    `json:"starts_on"`
	UserID           int64     `json:"user_id"`
	UserFullName     string    `json:"user_full_name"`
	UserEmail        string    `json:"user_email"`
	Status           Status    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Slot — состояние соревнования, от которого зависит, можно ли подать или одобрить заявку.
type Slot struct {
	CompetitionID int64
	StartsOn      time.Time
	Capacity      int
	Approved      int
}

func (s Slot) FreePlaces() int {
	return s.Capacity - s.Approved
}

// Filter — фильтры списка заявок для администратора.
type Filter struct {
	Status        Status
	CompetitionID int64
}
