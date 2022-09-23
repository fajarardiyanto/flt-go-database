package interfaces

type SQLConfig struct {
	Enable            bool   `yaml:"enable"`
	Driver            string `yaml:"driver"`
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	Database          string `yaml:"database"`
	Options           string `yaml:"options"`
	Connection        string `yaml:"connection"`
	AutoReconnect     bool   `yaml:"autoReconnect"`
	StartInterval     int    `yaml:"startInterval"`
	MaxError          int    `yaml:"maxError"`
	Sslmode           string `yaml:"sslmode"`
	TimeoutConnection int    `yaml:"timeoutConnection"`
}

type ElasticSearchProviderConfig struct {
	Enable        bool   `yaml:"enable"`
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	Password      string `yaml:"password"`
	Username      string `yaml:"username"`
	IndexName     string `yaml:"indexName"`
	MaxError      int    `yaml:"maxError"`
	AutoReconnect bool   `yaml:"autoReconnect"`
	StartInterval int    `yaml:"startInterval"`
}

type RedisProviderConfig struct {
	Enable        bool   `yaml:"enable"`
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	Password      string `yaml:"password"`
	AutoReconnect bool   `yaml:"autoReconnect"`
	StartInterval int    `yaml:"startInterval"`
	MaxError      int    `yaml:"maxError"`
}

type MongoProviderConfig struct {
	Enable            bool   `yaml:"enable"`
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	AutoReconnect     bool   `yaml:"autoReconnect"`
	MaxError          int    `yaml:"maxError"`
	StartInterval     int    `yaml:"startInterval"`
	TimeoutConnection int    `yaml:"timeoutConnection"`
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
