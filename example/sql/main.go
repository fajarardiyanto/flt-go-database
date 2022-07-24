package main

import (
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib"
	log "gitlab.com/fajardiyanto/flt-go-logger/lib"
)

func main() {
	logger := log.NewLib().Init()
	logger.SetFormat("text").SetLevel("debug")

	db := lib.NewLib()
	db.Init(logger)

	mysql := db.LoadSQLDatabase(interfaces.SQLConfig{
		Enable:        true,
		Driver:        "mysql",
		Host:          "127.0.0.1",
		Port:          3306,
		Username:      "root",
		Password:      "root",
		Database:      "mysql",
		AutoReconnect: true,
		StartInterval: 2,
	})
	if err := mysql.LoadSQL(); err != nil {
		fmt.Println(err)
		return
	}

	var version string
	err := mysql.Orm().Raw(`SELECT @@VERSION`).Scan(&version).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(version)
}
