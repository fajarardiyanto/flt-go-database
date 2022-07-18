package elasticsearch

import (
	"encoding/json"
	"fmt"
)

func (c *ElasticSearch) Delete(id string) error {
	res, err := c.elastic.Delete(c.config.IndexName, id)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return err
		}

		return fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	return nil
}
