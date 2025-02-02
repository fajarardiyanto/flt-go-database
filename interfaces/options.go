package interfaces

import "context"

type SQLConfig struct {
	Enable            bool   `yaml:"enable" default:"false"`
	Driver            string `yaml:"driver" default:"mysql"`
	Host              string `yaml:"host" default:"127.0.0.1"`
	Port              int    `yaml:"port" default:"3306"`
	Username          string `yaml:"username" default:"root"`
	Password          string `yaml:"password" default:"root"`
	Database          string `yaml:"database" default:"mydb"`
	Options           string `yaml:"options" default:""`
	Connection        string `yaml:"connection" default:""`
	AutoReconnect     bool   `yaml:"autoReconnect" default:"false"`
	StartInterval     int    `yaml:"startInterval" default:"5"`
	MaxError          int    `yaml:"maxError" default:"5"`
	Sslmode           string `yaml:"sslmode" default:"false"`
	TimeoutConnection int    `yaml:"timeoutConnection" default:"3000"`
	CustomPool        bool   `yaml:"customPool" default:"5"`
	MaxConn           int    `yaml:"maxConn" default:"5"`
	MaxIdle           int    `yaml:"maxIdle" default:"5"`
	LifeTime          int    `yaml:"lifeTime" default:"5"`
}

type ElasticSearchProviderConfig struct {
	Enable        bool   `yaml:"enable" default:"false"`
	Host          string `yaml:"host" default:"127.0.0.1"`
	Port          int    `yaml:"port" default:"9200"`
	Password      string `yaml:"password" default:"root"`
	Username      string `yaml:"username" default:"root"`
	IndexName     string `yaml:"indexName" default:"mydb_idx"`
	MaxError      int    `yaml:"maxError" default:"5"`
	AutoReconnect bool   `yaml:"autoReconnect" default:"false"`
	StartInterval int    `yaml:"startInterval" default:"5"`
}

type RedisProviderConfig struct {
	Enable        bool   `yaml:"enable" default:"false"`
	Host          string `yaml:"host" default:"127.0.0.1"`
	Port          int    `yaml:"port" default:"6379"`
	Password      string `yaml:"password" default:""`
	AutoReconnect bool   `yaml:"autoReconnect" default:"false"`
	StartInterval int    `yaml:"startInterval" default:"5"`
	MaxError      int    `yaml:"maxError" default:"5"`
}

type MongoProviderConfig struct {
	Enable            bool   `yaml:"enable" default:"false"`
	Host              string `yaml:"host" default:"127.0.0.1"`
	Port              int    `yaml:"port" default:"27017"`
	Username          string `yaml:"username" default:"root"`
	Password          string `yaml:"password" default:"root"`
	AutoReconnect     bool   `yaml:"autoReconnect" default:"false"`
	MaxError          int    `yaml:"maxError" default:"5"`
	StartInterval     int    `yaml:"startInterval" default:"5"`
	TimeoutConnection int    `yaml:"timeoutConnection" default:"3000"`
}

type KafkaProviderConfig struct {
	Enable           bool   `yaml:"enable" default:"false"`
	Host             string `yaml:"host" default:"127.0.0.1:9092"`
	Registry         string `yaml:"registry" default:""`
	Username         string `yaml:"username" default:""`
	Password         string `yaml:"password" default:""`
	SecurityProtocol string `yaml:"securityProtocol" default:"SASL_SSL"`
	Mechanisms       string `yaml:"mechanisms" default:"PLAIN"`
	Debug            string `yaml:"debug" default:"consumer"`
}

type RabbitMQProviderConfig struct {
	Enable              bool   `yaml:"enable" default:"false"`
	Host                string `yaml:"host" default:"127.0.0.1"`
	Port                int    `yaml:"port" default:"5672"`
	Username            string `yaml:"username" default:"guest"`
	Password            string `yaml:"password" default:"guest"`
	ReconnectDuration   int    `yaml:"reconnectDuration" default:"5"`
	DedicatedConnection bool   `yaml:"dedicatedConnection" default:"false"`
}

type RabbitMQOptions struct {
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Durable      bool
	AutoDeleted  bool
	NoWait       bool
	Encoding     Encoding
}

type KafkaOptions struct {
	Topic          string
	RegistryValue  string
	Group          string
	SchemeVersion  int
	SchemeID       int
	Encoding       Encoding
	MultipleThread bool
	Limiter        int
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

type Encoding int

const (
	EncodingBase64Gob Encoding = iota
	EncodingGob
	EncodingProto
	EncodingNone
	EncodingJSON
)

type Messages interface {
	Exchange() string
	RoutingKey() string
	Decode(interface{}) error
	SetContext(context.Context)
	Context() context.Context
}

type EmbeddedOptions struct {
	Directory string
}
