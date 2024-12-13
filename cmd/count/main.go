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

func (dp *DatabaseProvider) SelectQuery() (int, error) {
	var c int
	row := dp.db.QueryRow("SELECT c FROM counter LIMIT 1")
	err := row.Scan(&c)
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (dp *DatabaseProvider) InsertQuery(c int) error {
	_, err := dp.db.Exec("INSERT INTO counter (c) VALUES ($1)", c)
	return err
}

func (dp *DatabaseProvider) SetQuery(c int) error {
	_, err := dp.db.Exec("UPDATE counter SET c=$1", c)
	return err
}

func (dp *DatabaseProvider) ClearQuery() error {
	_, err := dp.db.Exec("DELETE FROM counter")
	return err
}

func (h *Handlers) GetCounter(c echo.Context) error {
	counter, _ := h.dbProvider.SelectQuery()
	return c.String(http.StatusOK, fmt.Sprintf("Счётчик сейчас %d", counter))
}

func (h *Handlers) PostCounter(c echo.Context) error {
	counter, _ := h.dbProvider.SelectQuery()
	counter += 1
	var err error

	if counter > 1 {
		err = h.dbProvider.SetQuery(counter)
	} else {
		err = h.dbProvider.InsertQuery(counter)
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handlers) SetCounter(c echo.Context) error {
	num := c.QueryParam("num")
	var int_num int
	var err error

	if num == "" {
		int_num, _ = h.dbProvider.SelectQuery()
	} else {
		int_num, err = strconv.Atoi(num)
	}

	if err != nil {
		return c.String(http.StatusBadRequest, "num должен быть целым числом")
	}

	err = h.dbProvider.SetQuery(int_num)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, fmt.Sprintf("Значение %d установлено", int_num))
}

func (h *Handlers) ClearCounter(c echo.Context) error {
	err := h.dbProvider.ClearQuery()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "Счетик сброшен...")
}

func main() {
	address := flag.String("address", "127.0.0.1:8083", "Адрес для запуска сервиса Counter")
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
	e.GET("/get", h.GetCounter)
	e.GET("/post", h.PostCounter)
	e.GET("/clear", h.ClearCounter)
	e.GET("/set", h.SetCounter)

	// Запускаем веб-сервер на указанном адресе
	if err = e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
