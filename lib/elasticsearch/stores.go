package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func (c *ElasticSearch) Create(index string, id string, values interface{}) error {
	payload, err := json.Marshal(values)
	if err != nil {
		c.log.Error(err)
		return err
	}

	ctx := context.Background()
	res, err := esapi.CreateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(payload),
	}.Do(ctx, c.elastic)
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
