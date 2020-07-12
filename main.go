package main

import (
	"context"
	"log"
	"strconv"
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
	orderCollection := client.Database("teleshop").Collection("order")

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
		btnPlaceAnOrder    = menu.Text("Place an order")
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
		b.Send(m.Sender, "Welcome to the teleshop:", menu)
	})

	b.Handle(&btnMainMenu, func(m *tb.Message) {
		menu.Reply(
			menu.Row(btnMenu, btnBag),
		)
		b.Send(m.Sender, "Welcome to the teleshop:", menu)
	})

	b.Handle(&btnMenu, func(m *tb.Message) {
		menu.Reply(
			menu.Row(btnSushi, btnPizza, btnDesert),
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Choose a category of food you wanna buy:", menu)
	})

	b.Handle(&btnSushi, func(m *tb.Message) {
		menu.Reply(
			menu.Row(btnPhiladelphia, btnUnagiPhila),
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Now choose a food you like:", menu)
	})

	b.Handle(&btnPhiladelphia, func(m *tb.Message) {
		var position Position
		filterPosition := bson.D{{"name", "Philadelphia"}, {"category", "🍣 Sushi 🍣"}}
		_ = positionCollection.FindOne(context.TODO(), filterPosition).Decode(&position)
		p := &tb.Photo{File: tb.FromURL(position.Src)}
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
		var reply string
		filterCustomer := bson.D{{"username", m.Sender.Username}, {"telegramID", m.Sender.ID}}
		_ = customerCollection.FindOne(context.TODO(), filterCustomer).Decode(&customer)
		if len(customer.Bag) <= 0 {
			reply = "Your bag is empty!"
			menu.Reply(
				menu.Row(btnMainMenu),
			)
		} else {
			var position Position
			reply = "Positions in your bag:\n"
			totalPrice := 0
			for _, positionID := range customer.Bag {
				filterPosition := bson.D{{"_id", positionID}}
				_ = positionCollection.FindOne(context.TODO(), filterPosition).Decode(&position)
				reply += position.Name + ": " + strconv.Itoa(position.Price) + " " + position.Currency + "\n"
				totalPrice += position.Price
			}
			reply += "Total price: " + strconv.Itoa(totalPrice)
			menu.Reply(
				menu.Row(btnClear, btnPlaceAnOrder),
				menu.Row(btnMainMenu),
			)
		}
		b.Send(m.Sender, reply, menu)
	})

	b.Handle(&btnPlaceAnOrder, func(m *tb.Message) {
		var customer Customer
		filter := bson.D{{"username", m.Sender.Username}, {"telegramID", m.Sender.ID}}
		_ = customerCollection.FindOne(context.TODO(), filter).Decode(&customer)

		// Create an order
		_, _ = orderCollection.InsertOne(context.TODO(), bson.D{
			{"date", time.Now()},
			{"customer", customer.ID},
			{"bag", customer.Bag},
		})

		// Clear the customer's bag
		filterCustomer := bson.D{{"_id", customer.ID}}
		updateCustomerBag := bson.D{
			{"$set", bson.D{{"bag", []primitive.ObjectID{}}}},
		}
		_, _ = customerCollection.UpdateOne(context.TODO(), filterCustomer, updateCustomerBag)

		menu.Reply(
			menu.Row(btnMainMenu),
		)
		b.Send(m.Sender, "Your order is placed!", menu)
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
