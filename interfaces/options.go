package interfaces

type ElasticSearchProviderConfig struct {
	Enable        bool   `yaml:"enable" default:"false" desc:"config:elasticsearch:enable"`
	Host          string `yaml:"host" default:"127.0.0.1" desc:"config:elasticsearch:host"`
	Port          int    `yaml:"port" default:"9200" desc:"config:elasticsearch:port"`
	Password      string `yaml:"password" default:"" desc:"config:elasticsearch:password"`
	Username      string `yaml:"username" default:"elastic" desc:"config:elasticsearch:username"`
	IndexName     string `yaml:"index_name" default:"elasticsearch_idx" desc:"config:elasticsearch:index_name"`
	MaxError      int    `yaml:"maxerror" default:"5"  desc:"config:elasticsearch:maxerror"`
	AutoReconnect bool   `yaml:"autoreconnect" default:"false"  desc:"config:elasticsearch:autoreconnect"`
	StartInterval int    `yaml:"startinterval" default:"2"  desc:"config:elasticsearch:startinterval"`
}

type ElasticSearchOptions struct {
	Size  int    `default:"25"`
	Query string `default:""`
	Sort  string `default:"asc"`
	After []string
}

type SearchResultsElasticSearch struct {
	Total int           `json:"total"`
	Hits  []interface{} `json:"hits"`
}
