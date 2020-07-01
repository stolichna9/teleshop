package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB!")
	customerCollection := client.Database("teleshop").Collection("customer")

	b, _ := tb.NewBot(tb.Settings{
		Token:  "",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	var (
		menu        = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
		btnMenu     = menu.Text("📖 Menu 📖")
		btnBag      = menu.Text("👜 Bag 👜")
		btnSushi    = menu.Text("🍣 Sushi 🍣")
		btnPizza    = menu.Text("🍕 Pizza 🍕")
		btnDesert   = menu.Text("🍰 Desert 🍰")
		btnMainMenu = menu.Text("↩ Main Menu ↩")
	)

	b.Handle("/start", func(m *tb.Message) {
		var customer Customer
		filterUser := bson.D{{"username", m.Sender.Username}, {"telegramID", m.Sender.ID}}
		err := customerCollection.FindOne(context.TODO(), filterUser).Decode(&customer)
		if err != nil {
			_, err = customerCollection.InsertOne(context.TODO(), bson.D{
				{"username", m.Sender.Username},
				{"telegramID", m.Sender.ID},
			})
		}
		menu.Reply(
			menu.Row(btnMenu, btnBag),
		)
		b.Send(m.Sender, "Welcome to the teleshop: ", menu)
	})

	b.Handle(&btnMainMenu, func(m *tb.Message) {
		menu.Reply(
			menu.Row(btnMenu, btnBag),
		)
		b.Send(m.Sender, "Welcome to the teleshop: ", menu)
	})

	b.Handle(&btnMenu, func(m *tb.Message) {
		menu.Reply(
			menu.Row(btnSushi, btnPizza, btnDesert),
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Choose a category of food you wanna buy: ", menu)
	})

	b.Start()
}
