package db

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type DBManager struct {
	db  *sql.DB
	mux sync.Mutex
}

func NewDBManager(dsn string) (*DBManager, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &DBManager{db: db}, nil
}

func (dm *DBManager) GetDB() *sql.DB {
	dm.mux.Lock()
	defer dm.mux.Unlock()
	return dm.db
}

func (dm *DBManager) Close() {
	dm.mux.Lock()
	defer dm.mux.Unlock()
	if dm.db != nil {
		dm.db.Close()
	}
}

func (dm *DBManager) BeginTransaction() (*sql.Tx, error) {
	return dm.GetDB().Begin()
}

func (dm *DBManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	return dm.GetDB().Exec(query, args...)
}

func (dm *DBManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return dm.GetDB().Query(query, args...)
}
