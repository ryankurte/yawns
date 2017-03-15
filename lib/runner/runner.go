/**
 * OpenNetworkSim Runner
 * The runner module is designed to manage client applications, allowing ONS to launch and maintain instances of a client
 * application to simplify the scripting and automation of simulations.
 *
 * Note that this module heavily relies on exec, which is appropriate for a client application but not for server side use.
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package runner

import ()

const (
	// DefaultRunCommand is the default command executed by the runner
	DefaultRunCommand = "{{server_address}} {{client_address}}"
)

type Runner struct {
}

func NewRunner() *Runner {
	return &Runner{}
}
