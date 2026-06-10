package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"

	"template_sch/internal/config"
)

type Connections struct {
	Primary   *sql.DB
	Secondary *sql.DB
}

func OpenConnections(ctx context.Context, cfg config.DatabaseConfig, logger *slog.Logger) (*Connections, error) {
	logger = logger.With("component", "database")

	primary, err := Open(ctx, cfg.Primary)
	if err != nil {
		return nil, err
	}

	secondary, err := Open(ctx, cfg.Secondary)
	if err != nil {
		if primary != nil {
			_ = primary.Close()
		}
		return nil, err
	}

	logConnected(logger, cfg.Primary, primary)
	logConnected(logger, cfg.Secondary, secondary)

	return &Connections{
		Primary:   primary,
		Secondary: secondary,
	}, nil
}

func Open(ctx context.Context, cfg config.DatabaseConnectionConfig) (*sql.DB, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if cfg.DSN == "" {
		return nil, fmt.Errorf("%s database dsn is required when database is enabled", cfg.Name)
	}

	driver, err := sqlDriverName(cfg.Driver)
	if err != nil {
		return nil, fmt.Errorf("%s database driver: %w", cfg.Name, err)
	}

	db, err := sql.Open(driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", cfg.Name, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping %s database: %w", cfg.Name, err)
	}

	return db, nil
}

func Close(conns *Connections) error {
	if conns == nil {
		return nil
	}

	var closeErr error
	if conns.Secondary != nil {
		if err := conns.Secondary.Close(); err != nil {
			closeErr = fmt.Errorf("close secondary database: %w", err)
		}
	}

	if conns.Primary != nil {
		if err := conns.Primary.Close(); err != nil {
			if closeErr != nil {
				return fmt.Errorf("%v; close primary database: %w", closeErr, err)
			}
			closeErr = fmt.Errorf("close primary database: %w", err)
		}
	}

	return closeErr
}

func sqlDriverName(driver string) (string, error) {
	switch strings.ToLower(driver) {
	case "postgres", "postgresql", "pgx":
		return "pgx", nil
	case "mysql":
		return "mysql", nil
	case "mssql", "sqlserver", "sql_server":
		return "sqlserver", nil
	default:
		return "", fmt.Errorf("unsupported driver %q, use postgres, mysql, or sqlserver", driver)
	}
}

func logConnected(logger *slog.Logger, cfg config.DatabaseConnectionConfig, db *sql.DB) {
	if db == nil {
		return
	}

	logger.Info("database connected", "name", cfg.Name, "driver", cfg.Driver)
}
