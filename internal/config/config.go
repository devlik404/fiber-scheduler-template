package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"template_sch/timezone"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Scheduler SchedulerConfig
	Task      TaskConfig
}

type AppConfig struct {
	Name         string
	Environment  string
	LogLevel     string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type SchedulerConfig struct {
	Enabled  bool
	Location string
}

type DatabaseConfig struct {
	Primary   DatabaseConnectionConfig
	Secondary DatabaseConnectionConfig
}

type DatabaseConnectionConfig struct {
	Name            string
	Enabled         bool
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
}

type TaskConfig struct {
	ExampleSchedule   string
	CaseSchedule      string
	ParseFileSchedule string
	// generator:task-config-field
}

func Load() (Config, error) {
	readTimeout, err := durationEnv("APP_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := durationEnv("APP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	idleTimeout, err := durationEnv("APP_IDLE_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}

	schedulerEnabled, err := boolEnv("SCHEDULER_ENABLED", true)
	if err != nil {
		return Config{}, err
	}

	databaseConfig, err := loadDatabaseConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		App: AppConfig{
			Name:         stringEnv("APP_NAME", "template-scheduler"),
			Environment:  stringEnv("APP_ENV", "development"),
			LogLevel:     stringEnv("APP_LOG_LEVEL", "info"),
			Port:         stringEnv("APP_PORT", "8080"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		Database: databaseConfig,
		Scheduler: SchedulerConfig{
			Enabled:  schedulerEnabled,
			Location: stringEnv("SCHEDULER_LOCATION", timezone.Default),
		},
		Task: TaskConfig{
			ExampleSchedule:   stringEnv("EXAMPLE_TASK_SCHEDULE", "*/1 * * * *"),
			CaseSchedule:      stringEnv("CASE_TASK_SCHEDULE", "*/1 * * * *"),
			ParseFileSchedule: stringEnv("PARSE_FILE_SCHEDULE", "*/5 * * * *"),
			// generator:task-config-load
		},
	}, nil
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	primary, err := loadDatabaseConnectionConfig("DB_PRIMARY", "primary", "postgres")
	if err != nil {
		return DatabaseConfig{}, err
	}

	secondary, err := loadDatabaseConnectionConfig("DB_SECONDARY", "secondary", "mysql")
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		Primary:   primary,
		Secondary: secondary,
	}, nil
}

func loadDatabaseConnectionConfig(prefix string, name string, defaultDriver string) (DatabaseConnectionConfig, error) {
	enabled, err := boolEnv(prefix+"_ENABLED", false)
	if err != nil {
		return DatabaseConnectionConfig{}, err
	}

	maxOpenConns, err := intEnv(prefix+"_MAX_OPEN_CONNS", 10)
	if err != nil {
		return DatabaseConnectionConfig{}, err
	}

	maxIdleConns, err := intEnv(prefix+"_MAX_IDLE_CONNS", 5)
	if err != nil {
		return DatabaseConnectionConfig{}, err
	}

	connMaxLifetime, err := durationEnv(prefix+"_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return DatabaseConnectionConfig{}, err
	}

	connMaxIdleTime, err := durationEnv(prefix+"_CONN_MAX_IDLE_TIME", 10*time.Minute)
	if err != nil {
		return DatabaseConnectionConfig{}, err
	}

	pingTimeout, err := durationEnv(prefix+"_PING_TIMEOUT", 5*time.Second)
	if err != nil {
		return DatabaseConnectionConfig{}, err
	}

	return DatabaseConnectionConfig{
		Name:            name,
		Enabled:         enabled,
		Driver:          stringEnv(prefix+"_DRIVER", defaultDriver),
		DSN:             stringEnv(prefix+"_DSN", ""),
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
		PingTimeout:     pingTimeout,
	}, nil
}

func (c Config) HTTPAddress() string {
	return ":" + c.App.Port
}

func stringEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}

	return duration, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid bool for %s: %w", key, err)
	}

	return parsed, nil
}

func intEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid int for %s: %w", key, err)
	}

	return parsed, nil
}
