package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // or any driver for testing

	"github.com/ldaidone/gomemo/memo"
	"github.com/ldaidone/gomemo/pkg/backends"
	_ "github.com/ldaidone/gomemo/pkg/backends/memory"
)

func mockSetupDB() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	_, _ = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
	_, _ = db.Exec(`INSERT INTO users (id, name) VALUES (1, 'Alice'), (2, 'Bob'), (3, 'Charlie')`)
	return db
}

func queryUserByID(db *sql.DB, id int) (any, error) {
	time.Sleep(300 * time.Millisecond) // simulate slow query
	row := db.QueryRow("SELECT name FROM users WHERE id = ?", id)
	var name string
	if err := row.Scan(&name); err != nil {
		return nil, err
	}
	return name, nil
}

func RunDBQuery() {
	db := mockSetupDB()
	defer db.Close()

	backend, _ := backends.NewBackend("memory") // or "redis"
	m := memo.New(
		memo.WithBackend(backend),
		memo.WithTTL(30*time.Second),
		memo.WithMetrics(true),
	)

	ctx := context.Background()

	getUser := func(id int) (string, error) {
		key := fmt.Sprintf("user:%d", id)
		v, err := m.Get(ctx, key, func() (any, error) {
			fmt.Printf("Querying DB for user %d...\n", id)
			return queryUserByID(db, id)
		})
		if err != nil {
			return "", err
		}
		return v.(string), nil
	}

	// First query → DB hit
	name, _ := getUser(2)
	fmt.Printf("User 2: %s\n", name)

	// Second query → cached
	name, _ = getUser(2)
	fmt.Printf("User 2 (cached): %s\n", name)

	// Third query → cached
	name, _ = getUser(2)
	fmt.Printf("User 3 (cached): %s\n", name)

	// Fourth query → cached?
	name, _ = getUser(2)
	fmt.Printf("User 4 (cached): %s\n", name)

	stats := m.Metrics().Snapshot()
	fmt.Printf("Hits=%d Misses=%d Requests=%d HitRatio=%.1f%%\n",
		stats.Hits, stats.Misses, stats.Requests, stats.HitRatio()*100)
}
