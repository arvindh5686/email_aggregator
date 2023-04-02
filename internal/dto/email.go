package dto

type Email struct {
	FromAddress string `json:"from" binding:"required"`
	ToAddress   string `json:"to" binding:"required"`
	FromName    string `json:"from_name" binding:"required"`
	ToName      string `json:"to_name" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
	Body        string `json:"body" binding:"required"`
}
