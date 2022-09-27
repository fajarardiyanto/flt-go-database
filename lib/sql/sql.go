package sql

import (
	"fmt"
	"time"

	"github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	"gorm.io/gorm"
)

type SQL struct {
	reconnectAt int
	db          *gorm.DB
	log         logger.Logger
	lastTimeout time.Duration
	config      *interfaces.SQLConfig
}

func NewSQL(log logger.Logger, config *interfaces.SQLConfig) interfaces.SQL {
	return &SQL{
		config: config,
		log:    log,
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
