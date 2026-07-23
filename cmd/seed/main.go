package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed data/names.txt
var namesRaw string

//go:embed data/adjectives.txt
var adjectivesRaw string

//go:embed data/animals.txt
var animalsRaw string

type createURLRequest struct {
	LongURL string `json:"long_url"`
	Alias   string `json:"alias,omitempty"`
}

type jobResult struct {
	duration time.Duration
	status   int
	errMsg   string
}

func main() {
	n := flag.Int("n", 100, "number of URLs to seed")
	workers := flag.Int("workers", 10, "concurrent workers")
	target := flag.String("target", "http://localhost:8000", "API base URL")
	maxConns := flag.Int("max-conns", 100, "max idle connections per host")
	flag.Parse()

	names := parseLines(namesRaw)
	adjectives := parseLines(adjectivesRaw)
	animals := parseLines(animalsRaw)

	slog.Info("starting seed",
		"n", *n,
		"workers", *workers,
		"target", *target,
		"max_conns", *maxConns,
		"combinations", len(adjectives)*len(names)*len(animals),
	)

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        *maxConns,
			MaxIdleConnsPerHost: *maxConns,
			MaxConnsPerHost:     *maxConns,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	jobs := make(chan createURLRequest, *n)
	results := make(chan jobResult, *n)

	var wg sync.WaitGroup

	for range *workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range jobs {
				results <- post(client, *target, req)
			}
		}()
	}

	go func() {
		for range *n {
			name := names[rand.IntN(len(names))]
			adj := adjectives[rand.IntN(len(adjectives))]
			animal := animals[rand.IntN(len(animals))]
			jobs <- createURLRequest{
				LongURL: fmt.Sprintf("https://www.%s.com/%s/%s", name, adj, animal),
				Alias:   fmt.Sprintf("%s-%s-%s", adj, name, animal),
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	start := time.Now()

	var success, failed int64
	var totalNanos int64
	const logEvery = 100

	// track failures by reason for the final summary
	errCounts := make(map[string]int)

	for r := range results {
		totalNanos += int64(r.duration)
		if r.status == http.StatusCreated {
			success++
		} else {
			failed++
			errCounts[r.errMsg]++
		}

		done := success + failed
		if done%logEvery == 0 {
			slog.Info("progress", "done", done, "success", success, "failed", failed)
		}
	}

	elapsed := time.Since(start)
	total := success + failed

	var avg time.Duration
	if total > 0 {
		avg = time.Duration(totalNanos / total)
	}

	slog.Info("seed complete",
		"total", total,
		"success", success,
		"failed", failed,
		"elapsed", elapsed.Round(time.Millisecond),
		"avg_latency", avg.Round(time.Millisecond),
		"req_per_sec", fmt.Sprintf("%.1f", float64(total)/elapsed.Seconds()),
	)

	if len(errCounts) > 0 {
		slog.Warn("failure breakdown")
		for reason, count := range errCounts {
			slog.Warn("failure", "reason", reason, "count", count)
		}
	}
}

func post(client *http.Client, target string, req createURLRequest) jobResult {
	body, err := json.Marshal(req)
	if err != nil {
		return jobResult{errMsg: "marshal: " + err.Error()}
	}

	start := time.Now()
	resp, err := client.Post(target+"/urls", "application/json", bytes.NewReader(body))
	duration := time.Since(start)

	if err != nil {
		return jobResult{duration: duration, errMsg: "network: " + err.Error()}
	}
	// Body must be fully read before Close so the Transport can reuse the connection.
	// Skipping this causes every request to drop the TCP connection → TIME_WAIT exhaustion.
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		reason := fmt.Sprintf("http_%d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		return jobResult{duration: duration, status: resp.StatusCode, errMsg: reason}
	}

	return jobResult{duration: duration, status: resp.StatusCode}
}

func parseLines(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), "\n")
	lines := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			lines = append(lines, p)
		}
	}
	return lines
}
