package interfaces

type SQLConfig struct {
	Enable            bool
	Driver            string
	Host              string
	Port              int
	Username          string
	Password          string
	Database          string
	Options           string
	Connection        string
	AutoReconnect     bool
	StartInterval     int
	MaxError          int
	Sslmode           string
	TimeoutConnection int
}

type ElasticSearchProviderConfig struct {
	Enable        bool
	Host          string
	Port          int
	Password      string
	Username      string
	IndexName     string
	MaxError      int
	AutoReconnect bool
	StartInterval int
}

type ElasticSearchOptions struct {
	Size  int
	Query string
	Sort  string
	After []string
}

type SearchResultsElasticSearch struct {
	Total int           `json:"total"`
	Hits  []interface{} `json:"hits"`
}
