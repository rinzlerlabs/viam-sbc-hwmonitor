package clocks

type clockSensor interface {
	StartUpdating() error
	Close()
	GetReadingMap() map[string]interface{}
	Name() string
}
