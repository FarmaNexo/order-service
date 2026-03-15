// pkg/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config representa toda la configuración del servicio
type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Redis       RedisConfig    `mapstructure:"redis"`
	AWS         AWSConfig      `mapstructure:"aws"`
	SQS         SQSConfig      `mapstructure:"sqs"`
	Log         LogConfig      `mapstructure:"log"`
	Services    ServicesConfig `mapstructure:"services"`
	Consumer    ConsumerConfig `mapstructure:"consumer"`
	Payment     PaymentConfig  `mapstructure:"payment"`
	Delivery    DeliveryConfig `mapstructure:"delivery"`
}

// ServicesConfig configuración de servicios externos
type ServicesConfig struct {
	CatalogService  ServiceEndpoint `mapstructure:"catalog_service"`
	PharmacyService ServiceEndpoint `mapstructure:"pharmacy_service"`
	UserService     ServiceEndpoint `mapstructure:"user_service"`
}

// ServiceEndpoint configuración de un servicio externo
type ServiceEndpoint struct {
	BaseURL string `mapstructure:"base_url"`
}

// ConsumerConfig configuración del consumidor SQS
type ConsumerConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	PollInterval      time.Duration `mapstructure:"poll_interval"`
	MaxMessages       int32         `mapstructure:"max_messages"`
	VisibilityTimeout int32         `mapstructure:"visibility_timeout"`
}

// PaymentConfig configuración de pagos
type PaymentConfig struct {
	MockEnabled bool    `mapstructure:"mock_enabled"`
	SuccessRate float64 `mapstructure:"success_rate"`
}

// DeliveryConfig configuración de delivery
type DeliveryConfig struct {
	BaseFee               float64 `mapstructure:"base_fee"`
	FreeDeliveryThreshold float64 `mapstructure:"free_delivery_threshold"`
}

// ServerConfig configuración del servidor HTTP
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig configuración de PostgreSQL
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	Schema          string        `mapstructure:"schema"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// JWTConfig configuración de JWT (solo validación, no generación)
type JWTConfig struct {
	Secret              string        `mapstructure:"secret"`
	AccessTokenDuration time.Duration `mapstructure:"access_token_duration"`
	Issuer              string        `mapstructure:"issuer"`
}

// RedisConfig configuración de Redis
type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

// GetAddr retorna la dirección host:port de Redis
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// AWSConfig configuración de AWS
type AWSConfig struct {
	Region   string `mapstructure:"region"`
	Endpoint string `mapstructure:"endpoint"`
}

// SQSConfig configuración de SQS
type SQSConfig struct {
	OrderEventsQueueURL    string `mapstructure:"order_events_queue_url"`
	CatalogEventsQueueURL  string `mapstructure:"catalog_events_queue_url"`
	PharmacyEventsQueueURL string `mapstructure:"pharmacy_events_queue_url"`
}

// LogConfig configuración de logging
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

// LoadConfig carga la configuración basada en el environment
func LoadConfig(environment string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")

	// Buscar el archivo config en múltiples rutas
	configFile := findConfigFile(environment)
	if configFile == "" {
		return nil, fmt.Errorf("config file config.%s.yaml not found", environment)
	}

	// Leer el archivo YAML como string
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configFile, err)
	}

	// Expandir variables de entorno ${VAR} en el YAML
	expanded := os.ExpandEnv(string(content))

	// Pasar el YAML expandido a Viper
	if err := v.ReadConfig(strings.NewReader(expanded)); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	config.Environment = environment

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func findConfigFile(environment string) string {
	filename := fmt.Sprintf("config.%s.yaml", environment)
	paths := []string{"./configs", "../configs", "../../configs"}

	for _, dir := range paths {
		path := fmt.Sprintf("%s/%s", dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	return nil
}

// GetDSN retorna el Data Source Name para PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsLocal() bool {
	return c.Environment == "local"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
