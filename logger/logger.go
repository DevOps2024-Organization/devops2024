package logger

import "go.uber.org/zap"

var Log *zap.Logger

func init() {
    // Initialize Zap logger
    var err error
    Log, err = zap.NewDevelopment()
    if err != nil {
        panic("failed to initialize logger: " + err.Error())
    }
    defer Log.Sync() // flushes buffer, if any
}