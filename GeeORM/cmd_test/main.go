package main

import (
	"GeeORM"
	"GeeORM/log"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	engine, _ := GeeORM.NewEngine("sqlite3", "./gee.db")
	defer engine.Close()

	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	msg := fmt.Sprintf("Execute success, %d rows affected", count)
	log.Info(msg)
}
