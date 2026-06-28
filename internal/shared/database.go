package shared

// import (
// 	"context"
// 	"database/sql"

// 	_ "modernc.org/sqlite"
// )

// type DB struct {
// 	*sql.DB
// }

// func NewDB(ctx context.Context, path string) (*DB, error) {
// 	db, err := sql.Open("sqlite", path)
// 	if err != nil {
// 		return &DB{}, err
// 	}
// 	tx, err := db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return &DB{}, err
// 	}
// 	_, err = tx.Exec(`
// 		CREATE TABLE IF NOT EXISTS users (
// 			id INTEGER PRIMARY KEY AUTOINCREMENT,
// 			username TEXT NOT NULL UNIQUE,
// 			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
// 			key TEXT,
// 			token TEXT
// 		);
// 		`)
// 	if err != nil {
// 		tx.Rollback()
// 		return &DB{}, nil
// 	}
// 	tx.Commit()
// 	return &DB{
// 		db,
// 	}, nil
// }
