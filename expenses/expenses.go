package expenses

import "github.com/lib/pq"

type Expense struct {
	ID     int            `gorm:"primary_key" json:"id"`
	Title  string         `gorm:"type:text" json:"title"`
	Amount float64        `gorm:"type:float" json:"amount"`
	Note   string         `gorm:"type:text" json:"note"`
	Tags   pq.StringArray `gorm:"type:text[]" json:"tags"`
}

type CreateRequestBody struct {
	Title  string   `json:"title" binding:"required"`
	Amount float64  `json:"amount" binding:"required"`
	Note   string   `json:"note" binding:"required"`
	Tags   []string `json:"tags" binding:"required"`
}
