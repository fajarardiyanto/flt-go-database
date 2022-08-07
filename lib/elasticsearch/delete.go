package elasticsearch

import (
	"encoding/json"
	"fmt"
)

func (c *ElasticSearch) Delete(id string) error {
	res, err := c.elastic.Delete(c.config.IndexName, id)
	if err != nil {
		c.log.Error(err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			c.log.Error(err)
			return err
		}

		er := fmt.Sprintf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		c.log.Error(er)
		return fmt.Errorf(er)
	}

	return nil
}
