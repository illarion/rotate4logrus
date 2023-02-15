package rotate4logrus_test

import (
	"context"
	"fmt"
	"github.com/illarion/rotate4logrus/v2"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

var tmpDir = os.TempDir()

func setup() (string, func()) {
	randomFileName := fmt.Sprintf("rotating_%s.log", strconv.Itoa(rand.Int()))
	return randomFileName, func() {
		files, err := os.ReadDir(tmpDir)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			if !strings.HasPrefix(file.Name(), randomFileName) {
				continue
			}
			err := os.Remove(path.Join(tmpDir, file.Name()))
			if err != nil {
				panic(err)
			}
		}
	}
}

func TestHook_Fire(t *testing.T) {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.SetOutput(io.Discard)

	ctx := context.Background()
	fileName, tearDown := setup()
	defer tearDown()

	hook, err := rotate4logrus.New(ctx, rotate4logrus.HookConfig{
		Levels:   logrus.AllLevels,
		FilePath: path.Join(tmpDir, fileName),
		Rotate:   5,
		Size:     16000,
		Mode:     0666,
	})

	if err != nil {
		t.Error(err)
	}

	log.Level = logrus.TraceLevel
	log.Hooks.Add(hook)
	log.Debug("1")
}

// TestHook_Rotate tests if the log file is rotated
// after reaching the size limit
func TestHook_Rotate(t *testing.T) {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.SetOutput(io.Discard)

	ctx := context.Background()
	fileName, tearDown := setup()
	defer tearDown()

	hook, err := rotate4logrus.New(ctx, rotate4logrus.HookConfig{
		Levels:   logrus.AllLevels,
		FilePath: path.Join(tmpDir, fileName),
		Rotate:   5,
		Size:     16000,
		Mode:     0600,
	})

	if err != nil {
		t.Error(err)
	}
	msgPrefixLen := len("time=\"2023-02-15T11:53:08+02:00\" level=debug msg=")
	n := 16000 / msgPrefixLen

	log.Level = logrus.TraceLevel
	log.Hooks.Add(hook)
	for i := 0; i < n+10; i++ {
		log.Debug("1")
	}

	//check that there are rotated files
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Error(err)
	}

	rotatedFiles := 0
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileName) {
			rotatedFiles++
		}
	}

	if rotatedFiles != 2 {
		t.Errorf("expected 2 rotated files, got %d", rotatedFiles)
	}

}

func TestHook_Pause(t *testing.T) {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.SetOutput(io.Discard)

	ctx := context.Background()
	fileName, tearDown := setup()
	defer tearDown()

	hook, err := rotate4logrus.New(ctx, rotate4logrus.HookConfig{
		Levels:   logrus.AllLevels,
		FilePath: path.Join(tmpDir, fileName),
		Rotate:   5,
		Size:     16000,
		Mode:     0600,
	})

	if err != nil {
		t.Error(err)
	}
	msgPrefixLen := len("time=\"2023-02-15T11:53:08+02:00\" level=debug msg=")
	n := 16000 / msgPrefixLen

	continueFn := hook.Pause()
	defer continueFn()

	log.Level = logrus.TraceLevel
	log.Hooks.Add(hook)
	for i := 0; i < n+10; i++ {
		log.Debug("1")
	}

	//check that there are rotated files
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Error(err)
	}

	rotatedFiles := 0
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileName) {
			rotatedFiles++
		}
	}

	if rotatedFiles != 1 {
		t.Errorf("expected 2 rotated files, got %d", rotatedFiles)
	}

}
