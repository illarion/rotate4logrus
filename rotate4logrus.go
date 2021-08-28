package rotate4logrus

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
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
	Size int64
	// File mode, like 0600
	Mode os.FileMode
}

type hook struct {
	cfg  *HookConfig
	file *os.File
	size int64
	m    sync.Mutex
}

func (h *hook) open() error {
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
	h.size = stat.Size()
	return nil
}

func New(cfg HookConfig) (logrus.Hook, error) {

	hook := &hook{
		cfg: &cfg,
	}

	err := hook.open()
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (h *hook) Levels() []logrus.Level {
	return h.cfg.Levels
}

func (h *hook) Fire(entry *logrus.Entry) error {
	bytes, err := entry.Bytes()

	if err != nil {
		return fmt.Errorf("Could not convert log entry to bytes: %w", err)
	}

	if h.cfg.Size == 0 {
		_, err := h.file.Write(bytes)
		if err != nil {
			return fmt.Errorf("Could not write to log file %s: %w", h.cfg.FilePath, err)
		}
		return nil
	}

	if h.size+int64(len(bytes)) > h.cfg.Size {
		err = h.rotate()

		if err != nil {
			return fmt.Errorf("Could not rotate log files: %w", err)
		}
	}

	n, err := h.file.Write(bytes)
	if err != nil {
		return fmt.Errorf("Could not write to log file %s: %w", h.cfg.FilePath, err)
	}

	h.size += int64(n)

	return nil
}

func (h *hook) rotate() error {
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
