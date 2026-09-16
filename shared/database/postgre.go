package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	commonconfig "commerce-platform/shared/config"
)

func NewPostgreClient(ctx context.Context, cfg commonconfig.DatabaseConfig) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	return pgxpool.New(ctx, connString)
}
