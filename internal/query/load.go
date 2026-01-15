package query

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/mkelk/ticks-melk/internal/tick"
)

// LoadTicksParallel loads all ticks from the issues directory with bounded concurrency.
func LoadTicksParallel(issuesDir string) ([]tick.Tick, error) {
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return nil, fmt.Errorf("read issues dir: %w", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(issuesDir, entry.Name()))
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}

	var (
		mu    sync.Mutex
		ticks []tick.Tick
		errCh = make(chan error, 1)
		wg    sync.WaitGroup
		jobs  = make(chan string)
	)

	worker := func() {
		defer wg.Done()
		for path := range jobs {
			item, err := readTickFile(path)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			mu.Lock()
			ticks = append(ticks, item)
			mu.Unlock()
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	for _, path := range paths {
		select {
		case jobs <- path:
		case err := <-errCh:
			close(jobs)
			wg.Wait()
			return nil, err
		}
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return ticks, nil
}

func readTickFile(path string) (tick.Tick, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return tick.Tick{}, fmt.Errorf("read tick: %w", err)
	}
	var t tick.Tick
	if err := json.Unmarshal(data, &t); err != nil {
		return tick.Tick{}, fmt.Errorf("parse tick: %w", err)
	}
	if err := t.Validate(); err != nil {
		return tick.Tick{}, fmt.Errorf("invalid tick: %w", err)
	}
	return t, nil
}
