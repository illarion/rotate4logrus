# Rotate4logrus

Logrotate-like hook for [sirupsen/logrus](https://github.com/sirupsen/logrus)

It produces files in the following format:

```{bash}
    rotating_log.txt
    rotating_log.txt.0
    rotating_log.txt.1
    rotating_log.txt.2
    rotating_log.txt.3
```

When there are more than `Rotate` files, the oldest one is deleted,
and the rest are renamed to the next number, so 
`rotating_log.txt.4` becomes `rotating_log.txt.3`,
`rotating_log.txt.3` becomes `rotating_log.txt.2`, and so on.

Rotating is done when the file size exceeds `Size` bytes.

## Example

```{go}
	var log = logrus.New()
  	log.Level = logrus.TraceLevel
	log.Formatter = new(logrus.TextFormatter)
	
	context := context.Background()

	hook, err := rotate4logrus.New(context, rotate4logrus.HookConfig{
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
	
	log.Info("Hello World!")
	
	continueFn := hook.Pause()
	
	log.Info("This line will never be rotated")
	log.Info("So that during this time you can, for example, make a zip archive of the log file")
	
	continueFn()
	
	log.Info("This line could be rotated")
```

# License

MIT. See LICENSE for more details
