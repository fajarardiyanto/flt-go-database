package main

import (
	"fmt"

	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib"
	log "github.com/fajarardiyanto/flt-go-logger/lib"
)

func main() {
	logger := log.NewLib()
	logger.Init("Modules SQL")

	db := lib.NewLib()
	db.Init(logger)

	cfg := new(interfaces.SQLConfig)
	mysql := db.LoadSQLDatabase(cfg)
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
