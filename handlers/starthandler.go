package handlers

import (
	"fmt"
	"sync"

	"gopkg.in/telebot.v3"
)

type User struct {
	PhoneNumber string
	FullName    string
	Location    string
}

var (
	userMap = make(map[int]*User)
	mu      sync.Mutex
)

func HandleStart(c telebot.Context, b *telebot.Bot) {
	userID := c.Message().Sender.ID

	// Ask for the first name
	b.Send(c.Message().Sender, "Ismingizni kiriting:")

	b.Handle(telebot.OnText, func(c telebot.Context) error {
		if c.Message().Sender.ID != userID {
			return nil
		}
		firstName := c.Text()

		// Ask for the last name
		b.Send(c.Message().Sender, "Familiyangizni kiriting:")

		b.Handle(telebot.OnText, func(c telebot.Context) error {
			if c.Message().Sender.ID != userID {
				return nil
			}
			lastName := c.Text()

			// Ask for the phone number
			markup := telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
			btnSharePhone := markup.Contact("Telefon raqamni yuborish")
			markup.Reply(markup.Row(btnSharePhone))

			b.Send(c.Message().Sender, "Telefon raqamingizni yuboring:", &markup)

			b.Handle(telebot.OnContact, func(c telebot.Context) error {
				if c.Message().Sender.ID != userID {
					return nil
				}
				phoneNumber := c.Message().Contact.PhoneNumber

				// Ask for the location
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
						PhoneNumber: phoneNumber,
						FullName:    firstName + " " + lastName,
						Location:    locationStr,
					}

					mu.Lock()
					userMap[int(userID)] = user
					mu.Unlock()

					btnWebApp := telebot.InlineButton{
						Text: "Open Web App",
						WebApp: &telebot.WebApp{
							URL: "https://067a-178-218-201-219.ngrok-free.app",
						},
					}

					inlineMarkup := telebot.ReplyMarkup{
						InlineKeyboard: [][]telebot.InlineButton{
							{btnWebApp},
						},
					}

					b.Send(c.Message().Sender, "Assalomu alaykum! "+firstName+" Eco Challenge botiga xush kelibsiz.", &inlineMarkup)
					return nil
				})
				return nil
			})
			return nil
		})
		return nil
	})
}
