package config

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConfig         *DBConfig
	ServerConfig     *ServerConfig
	JWTConfig        *JWTConfig
	MigrationsConfig *MigrationsConfig
	RedisConfig      *RedisConfig
}

type DBConfig struct {
	User            string
	Password        string
	DB              string
	Port            int
	Host            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	Signature     string
	TokenLifeSpan time.Duration
}

type MigrationsConfig struct {
	Path string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	dbConfig, err := newDBConfig()
	if err != nil {
		return nil, err
	}

	serverConfig, err := newServerConfig()
	if err != nil {
		return nil, err
	}

	JWTConfig, err := newJWTConfig()
	if err != nil {
		return nil, err
	}

	migrationsConfig, err := newMigrationsConfig()
	if err != nil {
		return nil, err
	}

	redisConfig, err := newRedisConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		DBConfig:         dbConfig,
		ServerConfig:     serverConfig,
		JWTConfig:        JWTConfig,
		MigrationsConfig: migrationsConfig,
		RedisConfig:      redisConfig,
	}, nil
}

func newDBConfig() (*DBConfig, error) {
	user, userExists := os.LookupEnv("POSTGRES_USER")
	password, passwordExists := os.LookupEnv("POSTGRES_PASSWORD")
	dbname, dbExists := os.LookupEnv("POSTGRES_DB")
	host, hostExists := os.LookupEnv("POSTGRES_HOST")
	portStr, portExists := os.LookupEnv("POSTGRES_PORT")

	if !userExists || !passwordExists || !dbExists || !hostExists || !portExists {
		return nil, errors.New("incomplete database configuration")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("invalid POSTGRES_PORT value")
	}

	return &DBConfig{
		User:            user,
		Password:        password,
		DB:              dbname,
		Port:            port,
		Host:            host,
		MaxOpenConns:    100,
		MaxIdleConns:    90,
		ConnMaxLifetime: 5 * time.Minute,
	}, nil
}

func newServerConfig() (*ServerConfig, error) {
	port, portExists := os.LookupEnv("SERVER_PORT")
	if !portExists {
		return nil, errors.New("SERVER_PORT is required")
	}

	return &ServerConfig{
		Port: port,
	}, nil
}

func newJWTConfig() (*JWTConfig, error) {
	signature, signatureExists := os.LookupEnv("JWT_SIGNATURE")
	if !signatureExists {
		return nil, errors.New("JWT_SIGNATURE is required")
	}

	lifespanStr, lifespanExists := os.LookupEnv("JWT_TOKEN_LIFESPAN")
	if !lifespanExists {
		return nil, errors.New("JWT_TOKEN_LIFESPAN is required")
	}

	lifespan, err := parseDurationWithDays(lifespanStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TOKEN_LIFESPAN value: %v", err)
	}

	return &JWTConfig{
		Signature:     signature,
		TokenLifeSpan: lifespan,
	}, nil
}

func parseDurationWithDays(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}

func newMigrationsConfig() (*MigrationsConfig, error) {
	path, pathExists := os.LookupEnv("MIGRATIONS_PATH")
	if !pathExists {
		return nil, errors.New("MIGRATIONS_PATH is required")
	}

	return &MigrationsConfig{
		Path: path,
	}, nil
}

func newRedisConfig() (*RedisConfig, error) {
	host, hostExists := os.LookupEnv("AUTH_REDIS_HOST")
	port, portExists := os.LookupEnv("AUTH_REDIS_PORT")
	password, passwordExists := os.LookupEnv("AUTH_REDIS_PASSWORD")
	dbStr, dbExists := os.LookupEnv("AUTH_REDIS_DB")

	if !hostExists || !portExists || !passwordExists || !dbExists {
		return nil, errors.New("incomplete Redis configuration")
	}

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return nil, errors.New("invalid AUTH_REDIS_DB value")
	}

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}, nil
}

func ConfigureDB(db *sql.DB, cfg *DBConfig) {
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
}
