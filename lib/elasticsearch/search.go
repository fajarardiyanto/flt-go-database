package elasticsearch

import (
	"encoding/json"
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"io"
	"strings"
)

func (c *ElasticSearch) Search(config interfaces.ElasticSearchOptions) (*interfaces.SearchResultsElasticSearch, error) {
	var result interfaces.SearchResultsElasticSearch

	q := c.BuildQuery(config)
	res, err := c.elastic.Search(
		c.elastic.Search.WithIndex(c.config.IndexName),
		c.elastic.Search.WithBody(q),
	)
	if err != nil {
		c.log.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		c.log.Error(fmt.Errorf("%s", "Index Not Found"))
		return &result, fmt.Errorf("%s", "Index Not Found")
	}

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			c.log.Error(err)
			return nil, err
		}

		c.log.Error(err)
		return nil, err
	}

	var r *EnvelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		c.log.Error(err)
		return nil, err
	}

	if len(r.Hits.Hits) < 1 {
		c.log.Error(err)
		return nil, fmt.Errorf("%s", "Data Not Found")
	}

	for _, h := range r.Hits.Hits {
		result.Total = r.Hits.Total.Value
		result.Hits = append(result.Hits, h.Source)
	}

	return &result, nil
}

func (c *ElasticSearch) BuildQuery(config interfaces.ElasticSearchOptions) io.Reader {
	var b strings.Builder

	b.WriteString("{\n")

	if config.Size == 0 {
		config.Size = 25
	}

	if config.Sort == "" {
		config.Sort = "asc"
	}

	if config.Query == "" {
		b.WriteString(fmt.Sprintf(searchAll, config.Size, config.Sort))
	} else {
		b.WriteString(fmt.Sprintf(searchMatch, config.Query, config.Size, config.Sort))
	}

	if len(config.After) > 0 && config.After[0] != "" && config.After[0] != "null" {
		b.WriteString(",\n")
		b.WriteString(fmt.Sprintf(`	"search_after": %s`, config.After))
	}

	b.WriteString("\n}")

	//fmt.Printf("%s\n", b.String())
	c.log.Info(b.String())
	return strings.NewReader(b.String())
}

const searchAll = `
	"query" : { "match_all" : {} },
	"size" : %d,
	"sort" : { "_doc" : "%s" }`

const searchMatch = `
	"query" : {
		"multi_match" : {
			"query" : %q,
			"operator" : "and"
		}
	},
	"highlight" : {
		"fields" : {
			"title" : { "number_of_fragments" : 0 },
			"alt" : { "number_of_fragments" : 0 },
			"transcript" : { "number_of_fragments" : 5, "fragment_size" : 25 }
		}
	},
	"size" : %d,
	"sort" : [ { "_score" : "desc" }, { "_doc" : "%s" } ]`
