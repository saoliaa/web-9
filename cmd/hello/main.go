package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "sandbox"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Обработчики HTTP-запросов
func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectQuery()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, msg)
}

func (h *Handlers) PostHello(c echo.Context) error {
	input := struct {
		Msg string `json:"msg"`
	}{}

	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err := h.dbProvider.InsertQuery(input.Msg)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusCreated, "")
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectQuery() (string, error) {
	var msg string

	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (dp *DatabaseProvider) InsertQuery(msg string) error {
	_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
	return err
}

func main() {
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервиса Hello")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	e := echo.New()

	// Регистрируем обработчики
	e.GET("/get", h.GetHello)
	e.POST("/post", h.PostHello)

	// Запускаем веб-сервер на указанном адресе
	if err = e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
