package db

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/it/base"
	"github.com/jmoiron/sqlx"
)

// PoolFn blah
type PoolFn func(url string) *sqlx.DB

// DB blah
type DB struct {
	connect PoolFn
	cfg     *base.Config
	con     *sqlx.DB
}

// New blah
func New(connectToDatabase PoolFn, cfg *base.Config) *DB {
	return &DB{
		cfg:     cfg,
		connect: connectToDatabase,
	}
}

// Init blah
func (db *DB) Init() *sqlx.DB {
	if db.cfg.CreateDB {
		createTestDatabase(db.cfg)
	}
	dbURL := buildDBURL(db.cfg)
	log.Println("Connecting to", dbURL)
	db.con = prepareTestDB(db.connect(dbURL), db.cfg)
	return db.con
}

// Shutdown blah
func (db *DB) Shutdown() {
	db.con.Close()
	if db.cfg.CreateDB {
		pgExec("DROP DATABASE "+db.cfg.DatabaseName, db.cfg)
	}
}

func buildDBURL(cfg *base.Config) string {
	pgURL := strings.Split(cfg.PostgresURL, "?")
	return pgURL[0] + "/" + cfg.DatabaseName + "?" + pgURL[1]
}

func createTestDatabase(cfg *base.Config) {
	cfg.DatabaseName = base.RandomStr()
	log.Println("Create-ing test database " + cfg.DatabaseName)
	pgExec("CREATE DATABASE "+cfg.DatabaseName, cfg)
}

func prepareTestDB(db *sqlx.DB, cfg *base.Config) *sqlx.DB {
	if cfg.MigrateDB {
		migrate(db, cfg)
	}
	return db
}

func migrate(db *sqlx.DB, cfg *base.Config) {
	log.Println("Migrate-ing database...")
	files, _ := filepath.Glob(filepath.Join(cfg.MigrationsFolder, "*.sql"))
	for _, file := range files {
		file, _ := os.Open(file)
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		var sql string
		for scanner.Scan() {
			sql += scanner.Text()
		}
		if _, err := db.Exec(sql); err != nil {
			log.Fatalln("Failed to exec ", sql, "on test DB")
		}
	}
}

func pgExec(stm string, cfg *base.Config) {
	db, err := sqlx.Connect("postgres", cfg.PostgresURL)
	if err != nil {
		log.Fatalln("Failed to open connection to", cfg.PostgresURL)
	}
	defer db.Close()
	if _, err = db.Exec(stm); err != nil {
		log.Fatalln("Failed to exec ", stm, "on", cfg.PostgresURL)
	}
}