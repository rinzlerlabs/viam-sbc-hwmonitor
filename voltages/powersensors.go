package voltages

type powerSensor interface {
	StartUpdating() error
	Close()
	GetReading() (voltage, current, power float64)
	GetReadingMap() map[string]interface{}
	GetName() string
}
