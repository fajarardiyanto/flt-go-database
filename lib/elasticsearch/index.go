package elasticsearch

import (
	"fmt"
	"strings"
)

func (c *ElasticSearch) CreateIndex(name string, mapping string) error {
	res, err := c.elastic.Indices.Create(name, c.elastic.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("error: %s", res)
	}
	return nil
}
