# Rotate4logrus

(Over)simplified logrotate-like hook for [sirupsen/logrus](https://github.com/sirupsen/logrus)

## Example

```
	var log = logrus.New()
  	log.Level = logrus.TraceLevel
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

	log.Hooks.Add(hook)
```

# License

MIT. See LICENSE for more details
