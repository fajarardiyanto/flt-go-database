package sql

import (
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func (c *SQL) MySQL() (err error) {
	if !c.config.Enable {
		return fmt.Errorf("aborted, database not enable in config, double check configuration again")
	}

	address := parsingMysqlURL(c.config)
	c.log.Info(fmt.Sprintf("Connecting to database mysql server %s@%s:%d", c.config.Username,
		c.config.Host, c.config.Port))

	if c.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       address,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		PrepareStmt: true,
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

func parsingMysqlURL(config *interfaces.SQLConfig) (connect string) {
	connect = config.Connection
	if len(connect) == 0 {
		connect = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.Username,
			config.Password,
			config.Host, config.Port,
			config.Database,
		)
		if len(config.Options) != 0 {
			connect += "?" + config.Options
		}
	}
	return connect
}
