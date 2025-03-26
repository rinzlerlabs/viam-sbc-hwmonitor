package powermanager

type PowerManager interface {
	ApplyPowerMode() (rebootRequired bool, err error)
	GetCurrentPowerMode() (powerMode interface{}, err error)
}
