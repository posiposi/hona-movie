package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/posiposi/hona-movie/backend/internal/config"
	"github.com/posiposi/hona-movie/backend/migrations"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

var commands = map[string]func(context.Context, *sql.DB) error{
	"up":     func(ctx context.Context, db *sql.DB) error { return goose.UpContext(ctx, db, ".") },
	"down":   func(ctx context.Context, db *sql.DB) error { return goose.DownContext(ctx, db, ".") },
	"status": func(ctx context.Context, db *sql.DB) error { return goose.StatusContext(ctx, db, ".") },
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}
	if _, ok := commands[os.Args[1]]; !ok {
		usage()
		os.Exit(1)
	}

	// 踏み台経由で本番へ適用する運用のため、DDL がメタデータロック待ちに入っても
	// Ctrl-C で中断できるようにする。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "usage: %s <up|down|status>\n", name)
	fmt.Fprintln(os.Stderr, "  down rolls back a single migration")
}

func run(ctx context.Context, command string) error {
	migrate, ok := commands[command]
	if !ok {
		return fmt.Errorf("unknown command: %s", command)
	}

	cfg, err := config.LoadDB()
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := migrate(ctx, db); err != nil {
		return fmt.Errorf("goose %s: %w", command, err)
	}
	return nil
}
