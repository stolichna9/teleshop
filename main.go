package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

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
	positionCollection := client.Database("teleshop").Collection("position")

	b, _ := tb.NewBot(tb.Settings{
		Token:  "",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	var (
		menu               = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
		btnMenu            = menu.Text("📖 Menu 📖")
		btnBag             = menu.Text("👜 Bag 👜")
		btnSushi           = menu.Text("🍣 Sushi 🍣")
		btnPhiladelphia    = menu.Text("Philadelphia")
		btnUnagiPhila      = menu.Text("Unagi Phila")
		btnPizza           = menu.Text("🍕 Pizza 🍕")
		btnDesert          = menu.Text("🍰 Desert 🍰")
		btnAddPhiladelphia = menu.Text("Add to my bag")
		btnClear           = menu.Text("Clear")
		btnMainMenu        = menu.Text("↩ Main Menu ↩")
	)

	b.Handle("/start", func(m *tb.Message) {
		var customer Customer
		filter := bson.D{{"username", m.Sender.Username}, {"telegramID", m.Sender.ID}}
		err := customerCollection.FindOne(context.TODO(), filter).Decode(&customer)
		if err != nil {
			_, err = customerCollection.InsertOne(context.TODO(), bson.D{
				{"username", m.Sender.Username},
				{"telegramID", m.Sender.ID},
				{"bag", []primitive.ObjectID{}},
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

	b.Handle(&btnSushi, func(m *tb.Message) {
		menu.Reply(
			menu.Row(btnPhiladelphia, btnUnagiPhila),
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Now choose a food you like: ", menu)
	})

	b.Handle(&btnPhiladelphia, func(m *tb.Message) {
		p := &tb.Photo{File: tb.FromURL("https://www.google.ru/images/branding/googlelogo/2x/googlelogo_color_92x30dp.png")}
		menu.Reply(
			menu.Row(btnAddPhiladelphia),
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, p, menu)
	})

	b.Handle(&btnAddPhiladelphia, func(m *tb.Message) {
		var position Position
		filterPosition := bson.D{{"name", "Philadelphia"}, {"category", "🍣 Sushi 🍣"}}
		_ = positionCollection.FindOne(context.TODO(), filterPosition).Decode(&position)
		filterCustomer := bson.D{{"username", m.Sender.Username}, {"telegramID", m.Sender.ID}}
		updateCustomerBag := bson.D{
			{"$push", bson.D{{"bag", position.ID}}},
		}
		_, _ = customerCollection.UpdateOne(context.TODO(), filterCustomer, updateCustomerBag)

		menu.Reply(
			menu.Row(btnAddPhiladelphia, btnBag),
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Position added to you bag!", menu)
	})

	b.Handle(&btnBag, func(m *tb.Message) {
		var customer Customer
		var position Position
		filterCustomer := bson.D{{"username", m.Sender.Username}, {"telegramID", m.Sender.ID}}
		_ = customerCollection.FindOne(context.TODO(), filterCustomer).Decode(&customer)
		for _, positionID := range customer.Bag {
			filterPosition := bson.D{{"_id", positionID}}
			_ = positionCollection.FindOne(context.TODO(), filterPosition).Decode(&position)
			println(position.Name)
		}

		menu.Reply(
			menu.Row(btnClear),
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Your bag: ", menu)
	})

	b.Handle(&btnClear, func(m *tb.Message) {
		filterCustomer := bson.D{{"username", m.Sender.Username}, {"telegramID", m.Sender.ID}}
		updateCustomerBag := bson.D{
			{"$set", bson.D{{"bag", []primitive.ObjectID{}}}},
		}
		_, _ = customerCollection.UpdateOne(context.TODO(), filterCustomer, updateCustomerBag)

		menu.Reply(
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Your bag is empty!", menu)
	})

	b.Start()
}
