package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/k0kubun/pp"
	_ "github.com/lib/pq"
	"gopkg.in/telebot.v3"
)

type User struct {
	Id          string
	PhoneNumber string
	FullName    string
	Location    string
	FirstName   string
	LastName    string
	Xp          int
	BirthDate   sql.NullString // Use sql.NullString for nullable fields
	Avatar      string
}

var (
	userMap = make(map[int]*User)
	mu      sync.Mutex
	db      *sql.DB
)

func init() {
	var err error
	connStr := "postgres://postgres:nodirbek@localhost:5432/ecodb?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleStart(c telebot.Context, b *telebot.Bot) {
	userID := c.Message().Sender.ID

	exists, err := userExists(int(userID))
	if err != nil {
		log.Println("Error checking user existence:", err)
		b.Send(c.Message().Sender, "Xatolik yuz berdi. Iltimos, qayta urinib ko'ring.")
		return
	}

	if exists {
		sendWebAppButton(c, b, int(userID))
		return
	}

	b.Send(c.Message().Sender, "Ismingizni kiriting:")

	b.Handle(telebot.OnText, func(c telebot.Context) error {
		if c.Message().Sender.ID != userID {
			return nil
		}

		firstName := c.Text()

		b.Send(c.Message().Sender, "Familiyangizni kiriting:")

		b.Handle(telebot.OnText, func(c telebot.Context) error {
			if c.Message().Sender.ID != userID {
				return nil
			}
			lastName := c.Text()

			markup := telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
			btnSharePhone := markup.Contact("Telefon raqamni yuborish")
			markup.Reply(markup.Row(btnSharePhone))

			b.Send(c.Message().Sender, "Telefon raqamingizni yuboring:", &markup)

			b.Handle(telebot.OnContact, func(c telebot.Context) error {
				if c.Message().Sender.ID != userID {
					return nil
				}
				phoneNumber := c.Message().Contact.PhoneNumber

				btnShareLocation := markup.Location("Joylashuvni yuborish")
				markup.Reply(markup.Row(btnShareLocation))

				b.Send(c.Message().Sender, "Joylashuvingizni yuboring:", &markup)

				b.Handle(telebot.OnLocation, func(c telebot.Context) error {
					if c.Message().Sender.ID != userID {
						return nil
					}
					location := c.Message().Location
					locationStr := fmt.Sprintf("Lat: %f, Lon: %f", location.Lat, location.Lng)

					user := &User{
						Id:          strconv.Itoa(int(userID)),
						PhoneNumber: phoneNumber,
						FullName:    firstName + " " + lastName,
						Location:    locationStr,
						FirstName:   firstName,
						LastName:    lastName,
						Xp:          5,
						BirthDate:   sql.NullString{},
						Avatar:      "https://media.rarebek.uz/avatars/3aa0c0e3-30bb-4ae8-bb79-d360572f2197.png",
					}

					err := insertUser(user)
					if err != nil {
						log.Println("Error inserting user:", err)
						b.Send(c.Message().Sender, "Xatolik yuz berdi. Iltimos, qayta urinib ko'ring.")
						return nil
					}

					sendWebAppButton(c, b, int(userID))
					return nil
				})
				return nil
			})
			return nil
		})
		return nil
	})
}

func userExists(userID int) (bool, error) {
	var exists bool
	query := `SELECT exists (SELECT 1 FROM users WHERE id=$1)`
	err := db.QueryRow(query, userID).Scan(&exists)
	return exists, err
}

func sendWebAppButton(c telebot.Context, b *telebot.Bot, userID int) {
	btnWebApp := telebot.InlineButton{
		Text: "Open Web App",
		WebApp: &telebot.WebApp{
			URL: "https://a3d4-185-213-229-5.ngrok-free.app",
		},
	}

	inlineMarkup := telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{btnWebApp},
		},
	}

	b.Send(c.Message().Sender, "Assalomu alaykum! Botga xush kelibsiz.", &inlineMarkup)
}

func insertUser(user *User) error {
	query := `
		INSERT INTO users (id, first_name, last_name, phone_number, location, xp, birth_date, avatar)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	pp.Println(user)
	_, err := db.Exec(query, user.Id, user.FirstName, user.LastName, user.PhoneNumber, user.Location, user.Xp, user.BirthDate, user.Avatar)
	return err
}
