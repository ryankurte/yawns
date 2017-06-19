package engine

import ()

// Medium interface defines a medium implementation for simulation purposes
type Medium interface {
	Send() chan interface{}
	Receive() chan interface{}
}

// Plugin interface defines functions required by plugin modules
type Plugin interface {
}
