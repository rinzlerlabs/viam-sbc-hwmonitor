package clocks

type ClockSensor interface {
	Close() error
	GetReadingMap() (map[string]interface{}, error)
	Name() string
}
