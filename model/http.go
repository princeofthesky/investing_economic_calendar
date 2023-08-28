package model

import "go-pinterest/db"

type EventResponse struct {
	Date     string       `json:"date"`
	Holidays []db.Holiday `json:"holidays"`
	Events   []db.EventInfo  `json:"events"`
}
type HttpResponse struct {
	Code    int `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data"`
}

type EventQuery struct {
	Countries  []int  `json:"countries"`
	Categories []int  `json:"categories"`
	From       string `json:"from"`
	To         string `json:"to"`
}
