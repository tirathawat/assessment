package expenses

import "github.com/lib/pq"

type Expense struct {
	ID     int            `gorm:"primary_key" json:"id" binding:"required"`
	Title  string         `gorm:"type:text" json:"title" binding:"required"`
	Amount float64        `gorm:"type:float" json:"amount" binding:"required"`
	Note   string         `gorm:"type:text" json:"note" binding:"required"`
	Tags   pq.StringArray `gorm:"type:text[]" json:"tags" binding:"required"`
}

type CreateRequestBody struct {
	Title  string   `json:"title" binding:"required"`
	Amount float64  `json:"amount" binding:"required"`
	Note   string   `json:"note" binding:"required"`
	Tags   []string `json:"tags" binding:"required"`
}
