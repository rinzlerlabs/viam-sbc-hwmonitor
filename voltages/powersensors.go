package voltages

type powerSensor interface {
	Close() error
	GetReading() (voltage, current, power float64, err error)
	GetReadingMap() (map[string]interface{}, error)
	GetName() string
}
