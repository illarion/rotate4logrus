package rotate4logrus

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
)

// HookConfig should be used in order to construct logrotating Hook
type HookConfig struct {
	// Levels to fire
	Levels []logrus.Level
	// FilePath full path to the log file
	FilePath string
	// Rotate log file count times before removing. If Rotate count is 0, old versions are removed rather than rotated.
	Rotate int
	// Log file is rotated only if it grows bigger then Size bytes. If Size is 0, ignored
	Size uint64
	// File mode, like 0600
	Mode os.FileMode
}

type Hook struct {
	ctx  context.Context
	cfg  *HookConfig
	file *os.File
	size uint64
	m    sync.Mutex

	pauses  chan pauseRequest
	queries chan pausedQuery
}

type pauseRequest struct {
	response chan func()
}

type pausedQuery struct {
	response chan bool
}

func New(context context.Context, cfg HookConfig) (*Hook, error) {

	hook := &Hook{
		ctx:     context,
		cfg:     &cfg,
		pauses:  make(chan pauseRequest),
		queries: make(chan pausedQuery),
	}

	err := hook.open()
	if err != nil {
		return nil, err
	}

	go func() {
		locks := make(map[*struct{}]struct{})
		for {
			select {
			case <-context.Done():
				return
			case r := <-hook.pauses:
				lock := &struct{}{}
				locks[lock] = struct{}{}
				r.response <- func() {
					delete(locks, lock)
				}
			case q := <-hook.queries:
				q.response <- len(locks) > 0
			}
		}
	}()

	return hook, nil
}

// Pause pauses the hook from rotating of the log file.
// It returns a function that should be called to resume the hook.
// It is recommended to call the returned function in a defer statement,
// or make sure that it is called as soon as possible, in order to avoid
// situations where the hook is paused for a long time so that log rotation
// is not performed.
func (h *Hook) Pause() (Continue func()) {
	select {
	case <-h.ctx.Done():
		return func() {}
	default:
	}

	r := pauseRequest{
		response: make(chan func()),
	}
	h.pauses <- r
	return <-r.response
}

func (h *Hook) paused() bool {
	select {
	case <-h.ctx.Done():
		return false
	default:
	}
	q := pausedQuery{
		response: make(chan bool),
	}
	h.queries <- q
	return <-q.response
}

func (h *Hook) open() error {
	select {
	case <-h.ctx.Done():
		return fmt.Errorf("Rotate4Logrus context is done")
	default:
	}
	file, err := os.OpenFile(h.cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, h.cfg.Mode)

	if err != nil {
		return fmt.Errorf("Could not open or create log file %s: %w", h.cfg.FilePath, err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("Could not get log file size %s: %w", h.cfg.FilePath, err)
	}

	h.file = file
	h.size = uint64(stat.Size())
	return nil
}

func (h *Hook) Levels() []logrus.Level {
	return h.cfg.Levels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	select {
	case <-h.ctx.Done():
		return fmt.Errorf("Rotate4Logrus context is done")
	default:
	}
	bytes, err := entry.Bytes()

	if err != nil {
		return fmt.Errorf("Could not convert log entry to bytes: %w", err)
	}

	paused := h.paused()

	if h.cfg.Size == 0 || paused {
		_, err := h.file.Write(bytes)
		if err != nil {
			return fmt.Errorf("Could not write to log file %s: %w", h.cfg.FilePath, err)
		}
		return nil
	}

	if h.size+uint64(len(bytes)) > h.cfg.Size {
		err = h.rotate()

		if err != nil {
			return fmt.Errorf("Could not rotate log files: %w", err)
		}
	}

	n, err := h.file.Write(bytes)
	if err != nil {
		return fmt.Errorf("Could not write to log file %s: %w", h.cfg.FilePath, err)
	}

	h.size += uint64(n)

	return nil
}

func (h *Hook) rotate() error {
	select {
	case <-h.ctx.Done():
		return fmt.Errorf("Rotate4Logrus context is done")
	default:
	}
	h.m.Lock()
	defer h.m.Unlock()

	err := h.file.Close()
	if err != nil {
		return fmt.Errorf("File %s already closed: %w", h.cfg.FilePath, err)
	}

	for k := h.cfg.Rotate - 1; k >= -1; k-- {
		n := k + 1

		filePath := h.cfg.FilePath
		if k >= 0 {
			filePath += "." + strconv.Itoa(k)
		}

		if n == h.cfg.Rotate {
			os.Remove(filePath)
			continue
		}

		newFilePath := h.cfg.FilePath + "." + strconv.Itoa(n)

		_, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		err = os.Rename(filePath, newFilePath)
		if err != nil {
			return fmt.Errorf("Could not rename %s to %s: %w", filePath, newFilePath, err)
		}
	}

	return h.open()

}
