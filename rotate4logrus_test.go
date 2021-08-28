package rotate4logrus_test

import (
	"testing"

	"github.com/illarion/rotate4logrus"
	"github.com/sirupsen/logrus"
)

func TestRotate(t *testing.T) {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)

	hook, err := rotate4logrus.New(rotate4logrus.HookConfig{
		Levels:   logrus.AllLevels,
		FilePath: "/tmp/rotating_log.txt",
		Rotate:   5,
		Size:     16000,
		Mode:     0600,
	})

	if err != nil {
		t.Error(err)
	}

	log.Level = logrus.TraceLevel

	log.Hooks.Add(hook)

	for {
		log.Debug("1")
	}

}
