package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	commonconfig "commerce-platform/shared/config"
)

func NewPostgreClient(ctx context.Context, cfg commonconfig.DatabaseConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolCfg.ConnConfig.Host = cfg.Host
	poolCfg.ConnConfig.Port = uint16(cfg.Port)
	poolCfg.ConnConfig.Database = cfg.Name
	poolCfg.ConnConfig.User = cfg.User
	poolCfg.ConnConfig.Password = cfg.Password
	poolCfg.ConnConfig.ConnectTimeout = cfg.Timeouts.Connect

	return pgxpool.NewWithConfig(ctx, poolCfg)
}
