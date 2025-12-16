package config

import (
	"os"
	"strconv"
)

type Config struct {
	APIPort           string
	APIHost           string
	PostgresHost      string
	PostgresPort      string
	PostgresUser      string
	PostgresPassword  string
	PostgresDB        string
	PostgresSSLMode   string
	RedisHost         string
	RedisPort         string
	RedisPassword     string
	RedisDB           string
	ShippingMode      string
	ShippingFeePercent float64
	FXUSDJPY          float64
	UserAgent         string
	RateLimitRPS      int
	RateLimitBurst    int
}

func Load() *Config {
	return &Config{
		APIPort:           getEnv("API_PORT", "8080"),
		APIHost:           getEnv("API_HOST", "0.0.0.0"),
		PostgresHost:      getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:      getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:      getEnv("POSTGRES_USER", "pricecompare"),
		PostgresPassword:  getEnv("POSTGRES_PASSWORD", "password"),
		PostgresDB:        getEnv("POSTGRES_DB", "pricecompare"),
		PostgresSSLMode:   getEnv("POSTGRES_SSLMODE", "disable"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           getEnv("REDIS_DB", "0"),
		ShippingMode:      getEnv("US_SHIP_MODE", "TABLE"),
		ShippingFeePercent: getFloatEnv("SHIPPING_FEE_PERCENT", 3.0),
		FXUSDJPY:          getFloatEnv("FX_USDJPY", 150.0),
		UserAgent:         getEnv("USER_AGENT", "PriceCompareBot/1.0"),
		RateLimitRPS:      getIntEnv("RATE_LIMIT_REQUESTS_PER_SECOND", 10),
		RateLimitBurst:    getIntEnv("RATE_LIMIT_BURST", 20),
	}
}

func (c *Config) DatabaseURL() string {
	return "postgres://" + c.PostgresUser + ":" + c.PostgresPassword +
		"@" + c.PostgresHost + ":" + c.PostgresPort + "/" + c.PostgresDB +
		"?sslmode=" + c.PostgresSSLMode
}

func (c *Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}

func (c *Config) ShippingConfig() ShippingConfig {
	return ShippingConfig{
		Mode:      c.ShippingMode,
		FeePercent: c.ShippingFeePercent,
		FXUSDJPY:  c.FXUSDJPY,
	}
}

type ShippingConfig struct {
	Mode       string
	FeePercent float64
	FXUSDJPY   float64
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

