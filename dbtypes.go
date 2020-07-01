package main

type Customer struct {
	Username   string `bson:"username"`
	TelegramID int    `bson:"telegramID"`
}
