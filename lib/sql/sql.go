package sql

import (
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "gitlab.com/fajardiyanto/flt-go-logger/interfaces"
	log "gitlab.com/fajardiyanto/flt-go-logger/lib"
	"gorm.io/gorm"
	"time"
)

type SQL struct {
	reconnectAt int
	db          *gorm.DB
	log         logger.Logger
	lastTimeout time.Duration
	config      interfaces.SQLConfig
}

func NewSQL(config interfaces.SQLConfig) interfaces.SQL {
	lo := log.NewLib().Init()
	lo.SetFormat("text").SetLevel("debug")

	return &SQL{
		config: config,
		log:    lo,
	}
}

func (c *SQL) Orm() *gorm.DB {
	return c.db
}

func (c *SQL) LoadSQL() error {
	switch c.config.Driver {
	case "mysql":
		return c.MySQL()
	case "postgresql":
		return c.PostgresSQL()

	}

	return fmt.Errorf("driver '%s' not supported", c.config.Driver)
}
