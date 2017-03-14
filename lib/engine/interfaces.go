package engine

// Connector interface defines methods required
// for modules that connect to devices
type Connector interface {
	Init(bindAddress string, h interface{}) error
	Send(address string, data []byte)
}

// Medium interface defines a medium implementation
// for simulation purposes
type Medium interface {
	// Check if the medium appears busy for a given device
	IsBusy(id uint) bool
	// Check whether a link exists between two devices
	CanSend(from uint, to uint) bool
}

// Plugin interface defines functions required by plugin modules
type Plugin interface {
}
