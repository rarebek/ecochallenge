package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"worker-bot/config"
	"worker-bot/handlers"
	"worker-bot/webhandlers"

	_ "worker-bot/docs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gopkg.in/telebot.v3"
)

type User struct {
	ID          string
	FullName    string
	PhoneNumber string
	Region      string
	LivingPlace string
}

func main() {
	connStr := "postgres://postgres:mubina2007@localhost:5432/ecochallengedb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pref := telebot.Settings{
		Token:  "7379288174:AAE45FbBl25Jrp52sRlG_HcLTBed-75ObYg",
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Bot started successfully.")

	b.Handle("/start", func(c telebot.Context) error {
		go handlers.HandleStart(c, b)
		return nil
	})

	go func() {
		b.Start()
	}()

	cfg := config.Load()

	psqlUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDatabase,
	)

	psqlConn, err := sqlx.Connect("postgres", psqlUrl)
	if err != nil {
		log.Fatalf("failed to connect to postgresql database: %v", err)
	}

	h := webhandlers.NewHandlerV1(psqlConn)

	r := gin.Default()

	r.GET("/questions", h.TestGenHandler)
	r.GET("/ranking", h.GetRanking)

	r.GET("/user/:id", h.GetUser)
	r.POST("/user", h.CreateUser)
	r.PUT("/user/:id", h.UpdateUser)
	r.DELETE("/user/:id", h.DeleteUser)
	r.GET("/users", h.ListUsers)

	r.GET("/event/:id", h.GetEvent)
	r.POST("/event", h.CreateEvent)
	r.PUT("/event/:id", h.UpdateEvent)
	r.DELETE("/event/:id", h.DeleteEvent)
	r.GET("/events", h.ListEvents)

	r.POST("/history", h.CreateHistory)
	r.GET("/history/:id", h.GetHistory)
	r.PUT("/history/:id", h.UpdateHistory)
	r.DELETE("/history/:id", h.DeleteHistory)
	r.GET("/history", h.ListHistory)

	url := ginSwagger.URL("swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start Gin server: ", err)
	}
}
