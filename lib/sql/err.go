package sql

import (
	"context"
	"strings"
	"time"
)

func (c *SQL) OnError(e error) (err error) {
	if c.config.MaxError == 0 {
		c.config.MaxError = 5
	}

	if c.reconnectAt >= c.config.MaxError {
		return e
	}

	c.reconnectAt++
	if !c.config.AutoReconnect {
		return e
	}

	var ttm = 2
	if c.lastTimeout.Seconds() == 0 {
		ttm = 1
		c.lastTimeout = time.Duration(c.config.StartInterval) * time.Second
	}

	c.lastTimeout = time.Duration(int(c.lastTimeout.Seconds())*ttm) * time.Second
	c.log.Error(e)
	c.log.Warnf("Reconnecting in %s", c.lastTimeout)
	time.Sleep(c.lastTimeout)

	return c.MySQL()
}

func (c *SQL) Close() {
	defer func() {
		if !c.config.AutoReconnect {
			c.db = nil
			c = nil
		}
	}()

	if sql, err := c.db.DB(); err == nil {
		if err := sql.Close(); err != nil {
			return
		}
		if _, err := sql.Conn(context.Background()); err != nil {
			if strings.Contains(err.Error(), "database is closed") {
				c.log.Debug("Database has been closed to %s@%s:%d",
					c.config.Username,
					c.config.Host, c.config.Port)
			}
		}

	}
	return
}
