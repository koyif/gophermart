package logger

import "go.uber.org/zap"

var Log = zap.NewNop()

func Initialize() error {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	zapLogger, err := config.Build()
	if err != nil {
		return err
	}

	Log = zapLogger
	return nil
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Float(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}
