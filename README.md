## config
Caches configuration objects for optional configurations that are stored by configuration type.
Configurations can be loaded from 
1. environment variables
2. config files

- ***config_man.go*** provides an object that caches Configger objects per type
- ***configger.go*** defines the object that ConfigMan stores
- ***slog_config.go*** convenience class that defines a Configger object to match how ***I*** want slog to work
  - uses [gopkg.in/natefinch/lumberjack.v2](gopkg.in/natefinch/lumberjack.v2) to rotate at
    - MaxSize of 50 MB
    - MaxAge of 30 days
    - MaxBackups of 30 logs 
