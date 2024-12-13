package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectQuery() (string, int, error) {
	var name string
	var age int

	row := dp.db.QueryRow("SELECT name, age FROM query ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&name, &age)
	if err != nil {
		return "", 0, err
	}

	return name, age, nil
}

func (dp *DatabaseProvider) InsertQuery(name string, age int) error {
	_, err := dp.db.Exec("INSERT INTO query (name, age) VALUES ($1, $2)", name, age)
	return err
}

func (dp *DatabaseProvider) ClearQuery() error {
	_, err := dp.db.Exec("DELETE FROM query")
	return err
}

// Обработчики HTTP-запросов
func (h *Handlers) GetQuery(c echo.Context) error {
	name, age, err := h.dbProvider.SelectQuery()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, fmt.Sprintf("Name=%s Age=%d", name, age))
}

func (h *Handlers) PostQuery(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		name = "Guest"
	}
	age := c.QueryParam("age")
	if age == "" {
		age = "0"
	}

	int_age, err := strconv.Atoi(age)
	if err != nil {
		return c.String(http.StatusBadRequest, "Age должен быть целым числом")
	}

	err = h.dbProvider.InsertQuery(name, int_age)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusCreated, "")
}

func (h *Handlers) ClearQuery(c echo.Context) error {
	err := h.dbProvider.ClearQuery()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "База данных очищена...")
}

func main() {
	address := flag.String("address", "127.0.0.1:8082", "Адрес для запуска сервиса Query")
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
	e.GET("/get", h.GetQuery)
	e.GET("/post", h.PostQuery)
	e.GET("/clear", h.ClearQuery)

	// Запускаем веб-сервер на указанном адресе
	if err = e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
