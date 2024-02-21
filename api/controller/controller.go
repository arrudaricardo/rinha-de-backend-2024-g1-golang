package controller

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"log"
	"rinha-de-backend-2024-q1-golang/database"
	"time"
)

type Transacoe struct {
	Valor     int    `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type TransacoeRow struct {
	Transacoe
	RealizadaEm time.Time `json:"realizada_em "`
}

func PostTransacoes(c fiber.Ctx) error {

	id, error := c.ParamsInt("id")

	if error != nil {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if id <= 0 || id > 5 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	transacoe := new(Transacoe)
	err := c.Bind().JSON(transacoe)
	if err != nil {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if len(transacoe.Descricao) > 10 || len(transacoe.Descricao) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if transacoe.Tipo != "d" && transacoe.Tipo != "c" {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	tx, err := database.ConnPool.Begin(context.Background())
	if err != nil {
		log.Println("Error while acquiring connection from the database pool!!", err.Error())
		return c.SendStatus(fiber.StatusInternalServerError)
	}

  _, err = tx.Exec(context.Background(), `SET LOCAL lock_timeout = '6s'`)
  _, err = tx.Exec(context.Background(), `SELECT pg_advisory_xact_lock($1)`, id)

	if err != nil {
		log.Println(err.Error())
		tx.Commit(context.Background())
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if err != nil {
		log.Println(err.Error())
		tx.Commit(context.Background())
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var saldo int
	var limite int
	err = tx.QueryRow(context.Background(), `SELECT s, l FROM a where i = $1`, id).Scan( &saldo, &limite)
	if err != nil {
		// log.Println(err.Error())
		tx.Commit(context.Background())
		return c.SendStatus(fiber.StatusNotFound)
	}


	newSaldo := saldo
	if transacoe.Tipo == "d" {
		newSaldo -= transacoe.Valor
	} else {
		newSaldo += transacoe.Valor
	}

	if newSaldo < -limite {
		tx.Commit(context.Background())
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	_, err = tx.Exec(context.Background(), `
	     INSERT INTO t (a, v, t, d, r)
	     VALUES ($1, $2, $3, $4, $5)
	   `, id, transacoe.Valor, transacoe.Tipo, transacoe.Descricao, time.Now())

	if err != nil {
		log.Println(err.Error())
		tx.Rollback(context.Background())
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	_, err = tx.Exec(context.Background(), `
	     UPDATE a SET s = $1 WHERE i = $2
	   `, newSaldo, id)

	if err != nil {
		log.Println(err.Error())
		tx.Rollback(context.Background())
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	tx.Commit(context.Background())
  

	return c.JSON(fiber.Map{"limite": limite, "saldo": newSaldo})
}

func GetExtrato(c fiber.Ctx) error {

	id, error := c.ParamsInt("id")

	if error != nil {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	tx, err := database.ConnPool.Acquire(context.Background())
	if err != nil {
		log.Println("Error while acquiring connection from the database pool!!", err.Error())
		return c.SendStatus(fiber.StatusInternalServerError)
	}

  defer tx.Release()

	var saldo int
	var limite int
	err = tx.QueryRow(context.Background(), `SELECT s, l FROM a where i = $1`, id).Scan(&saldo, &limite)

	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}


	rows, err := tx.Query(context.Background(), `SELECT t.v, t.t, t.d, t.r FROM t WHERE t.a = $1 ORDER BY r DESC LIMIT 10`, id)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	defer rows.Close()

	var rowSlice []TransacoeRow
	for rows.Next() {
		var r TransacoeRow
		rows.Scan(&r.Valor, &r.Tipo, &r.Descricao, &r.RealizadaEm)
		rowSlice = append(rowSlice, r)
	}


	return c.JSON(fiber.Map{
		"saldo": fiber.Map{
			"total":        saldo,
			"data_extrato": time.Now(),
			"limite":       limite,
		},
		"ultimas_transacoes": rowSlice,
	})

}
