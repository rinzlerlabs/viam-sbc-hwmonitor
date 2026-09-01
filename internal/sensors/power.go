package sensors

type PowerSensor interface {
	Close() error
	GetReading() (voltage, current, power float64, err error)
	GetReadingMap() (map[string]any, error)
	GetName() string
}
