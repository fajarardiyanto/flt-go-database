package sql

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func (c *SQL) PostgresSQL() (err error) {
	if !c.config.Enable {
		return fmt.Errorf("aborted, database not enable in config, double check configuration again")
	}

	address := c.parsingPostgresSQL()
	c.log.Debugf("Connecting to database postgresSQL server %s@%s:%d", c.config.Username,
		c.config.Host, c.config.Port)

	if c.db, err = gorm.Open(postgres.Open(address), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	}); err != nil {
		c.Close()
		return c.OnError(fmt.Errorf("%s", err.Error()))
	}

	return nil
}

func (c *SQL) parsingPostgresSQL() (connect string) {
	config := c.config
	connect = config.Connection
	if len(connect) == 0 {
		connect = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s connect_timeout=%d sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database,
			int(time.Duration(config.TimeoutConnection)*time.Millisecond), config.Sslmode)
		if len(config.Options) != 0 {
			connect += " " + config.Options
		}
	}
	return connect
}
