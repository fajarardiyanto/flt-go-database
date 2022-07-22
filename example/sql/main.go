package main

import (
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib/sql"
	"log"
)

func main() {
	mysql := sql.NewSQL(interfaces.SQLConfig{
		Enable:        true,
		Driver:        "mysql",
		Host:          "127.0.0.1",
		Port:          3334,
		Username:      "root",
		Password:      "root",
		Database:      "faltar",
		AutoReconnect: true,
		StartInterval: 2,
	})
	if err := mysql.LoadSQL(); err != nil {
		log.Println(err)
		return
	}

	var version string
	err := mysql.Orm().Raw(`SELECT @@VERSION`).Scan(&version).Error
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(version)
}
