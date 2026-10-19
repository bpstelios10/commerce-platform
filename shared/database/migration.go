package database

import (
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	commonconfig "commerce-platform/shared/config"
)

func RunMigrations(cfg commonconfig.DatabaseConfig, files fs.FS) error {
	source, err := iofs.New(files, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	query := url.Values{}
	query.Set("sslmode", "disable")
	query.Set("connect_timeout", strconv.Itoa(int(cfg.Timeouts.Connect.Seconds())))

	databaseURL := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     "/" + cfg.Name,
		RawQuery: query.Encode(),
	}).String()

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		source,
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("create migration: %w", err)
	}
	defer m.Close()

	// TODO i can test this if i create an interface but is it worth it?
	// type upper interface{ Up() error }
	//func applyMigrations(m upper) error {
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
