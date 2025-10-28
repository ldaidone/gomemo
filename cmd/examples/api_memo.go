package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ldaidone/gomemo/memo"
	"github.com/ldaidone/gomemo/pkg/backends"
	_ "github.com/ldaidone/gomemo/pkg/backends/memory"
	// _ "github.com/ldaidone/gomemo/pkg/backends/redis"
)

type Post struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

func fetchPost(id int) (any, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d", id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var p Post
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return p, nil
}

func RunAPIMemo() {
	backend, _ := backends.NewBackend("memory") // or "redis"
	m := memo.New(
		memo.WithBackend(backend),
		memo.WithTTL(1*time.Minute),
		memo.WithMetrics(true),
	)

	ctx := context.Background()

	// Memoized API call
	getPost := func(id int) (Post, error) {
		key := fmt.Sprintf("post-%d", id)
		v, err := m.Get(ctx, key, func() (any, error) {
			fmt.Printf("Fetching post %d from API...\n", id)
			return fetchPost(id)
		})
		if err != nil {
			return Post{}, err
		}
		return v.(Post), nil
	}

	// First call → fetch from API
	post, _ := getPost(1)
	fmt.Printf("First call: %v\n", post.Title)

	// Second call → instant cache hit
	post, _ = getPost(1)
	fmt.Printf("Second call (cached): %v\n", post.Title)

	stats := m.Metrics().Snapshot()
	fmt.Printf("Hits=%d Misses=%d Requests=%d HitRatio=%.1f%%\n",
		stats.Hits, stats.Misses, stats.Requests, stats.HitRatio()*100)
}
