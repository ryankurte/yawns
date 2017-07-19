package engine

import ()

// Medium interface defines a medium implementation for simulation purposes
type Medium interface {
	Send() chan interface{}
	Receive() chan interface{}
}
