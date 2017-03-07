package engine

// Connector interface defines methods required
// for modules that connect to devices
type Connector interface {
    Init(interface{}) error
}

type Medium interface {
    IsBusy(id uint) bool            // Check if the medium appears busy for a given device
    CanSend(from uint, to uint)     
}


