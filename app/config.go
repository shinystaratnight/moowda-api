package app

import (
	"fmt"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
	"os"
)

// Config stores the application-wide configurations
var Config appConfig

type appConfig struct {
	// the path to the error message file. Defaults to "config/errors.yaml"
	ErrorFile string `mapstructure:"error_file"`
	// the server port. Defaults to 8080
	ServerPort int `mapstructure:"server_port"`
	// the data source name (DSN) for connecting to the database. required.
	DSN string `mapstructure:"dsn"`
	// the data source name (DSN) for connecting to the test database. required.
	DSNTest string `mapstructure:"dsn_test"`
	// the signing method for JWT. Defaults to "HS256"
	JWTSigningMethod string `mapstructure:"jwt_signing_method"`
	// JWT signing key. required.
	JWTSigningKey string `mapstructure:"jwt_signing_key"`
	// JWT verification key. required.
	JWTVerificationKey string `mapstructure:"jwt_verification_key"`

	/////////////////////////////
	// Related to Emails
	/////////////////////////////

	// DefaultEmailAddress, the default sending email. required.
	DefaultEmailAddress string `mapstructure:"default_email_address"`
	// Sendgrid key for sending emails. required.
	SendgridAPIKey string `mapstructure:"sendgrid_api_key"`

	/////////////////////////////
	// Storage
	/////////////////////////////
	// Adapter: "s3" or "local"
	StorageAdapter string `mapstructure:"storage_adapter"`
	// Config must be a valid json map. Check Storage.Instance type functions in storage.s3 or storage.file for examples
	StorageConfig string `mapstructure:"storage_config"`
}

func (config appConfig) Validate() error {
	return validation.ValidateStruct(&config,
		validation.Field(&config.DSN, validation.Required),
		validation.Field(&config.DSNTest, validation.Required),
		validation.Field(&config.JWTSigningKey, validation.Required),
		validation.Field(&config.JWTVerificationKey, validation.Required),
		validation.Field(&config.DefaultEmailAddress, validation.Required),
		validation.Field(&config.SendgridAPIKey, validation.Required),
	)
}

// LoadConfig loads configuration from the given list of paths and populates it into the Config variable.
// The configuration file(s) should be named as development.yaml.
// Environment variables with the prefix "RESTFUL_" in their names are also read automatically.
func LoadConfig(configPaths ...string) error {
	v := viper.New()

	v.SetConfigName("development")

	if os.Getenv("ENV") != "" {
		v.SetConfigName(os.Getenv("ENV"))
	}

	v.SetConfigType("yaml")
	v.SetEnvPrefix("restful")
	v.AutomaticEnv()

	v.SetDefault("error_file", "config/errors.yaml")
	v.SetDefault("server_port", 8080)
	v.SetDefault("jwt_signing_method", "HS256")
	v.SetDefault("email_port", 587)

	// storage
	v.SetDefault("storage_adapter", "local")
	v.SetDefault("storage_config", `{"path": "static"}`)

	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read the configuration file: %s", err)
	}
	if err := v.Unmarshal(&Config); err != nil {
		return fmt.Errorf("failed to read the configuration file: %s", err)
	}

	return Config.Validate()
}
