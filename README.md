# Open Network Simulator

A Next-Generation Wireless Sensor Network Simulation Engine for Wireless Sensor Network (WSN) Research and Development.

This is designed to assist the development and evaluation of WSN operating systems, protocols, and applications in a programmable virtual environment, with a focus on replicating the radio environment of a physical environment in a virtual space.

This project consists of two parts. The core simulator (ons) that manages clients and connections, simulates the medium and environment, and processes Events to the simulation environment, and libyawns that implements a C Language radio interface for use in the node application.

## Status

[![Documentation](https://img.shields.io/badge/docs-godoc-blue.svg)](https://godoc.org/github.com/ryankurte/ons)
[![GitHub tag](https://img.shields.io/github/tag/ryankurte/ons.svg)](https://github.com/ryankurte/ons)
[![Build Status](https://travis-ci.com/ryankurte/ons.svg?token=s4CML2iJ2hd54vvqz5FP&branch=master)](https://travis-ci.com/ryankurte/ons/branches)

Very early prototype

## Goals

- [X] Event Engine
- [X] Common Medium interface
  - X ] Definition of wireless medium as appropriate for simulation tasks
- [X] Plugin support
  - [X]Standard Plugin Interface
- [X] PCAP streaming, file writing
- [X] Runnable / Client Management
- [ ] Platform / OS independent
  - [X] Generic Radio Driver
  - [ ] Cross compiled packages
- [ ] OpenGL / Map Visualisation

## Dependencies

Go compiler from  [golang.org/](https://golang.org/dl/).

- cmake
- sodium
- czmq
- protoc
- protobuf-c

### Debian / Ubuntu
```
sudo apt install libsodium-dev libzmq5-dev libczmq-dev libprotobuf-dev protobuf-compiler libprotobuf-c-dev protobuf-c-compiler pkg-config build-essential
```

### OSX
```
brew install libsodium czmq protobuf-c
```

## Building

1. `make tools` to fetch required (go) tools
2. `make deps` to update dependencies
3. `make` to build yawns

## Usage

ONS is designed to be platform and network agnostic. To simulate a given platform

1. Install ons
2. Create a wrapper for libyawns to adapt to the system under test
3. Create a simulation configuration file
4. Launch ons with the specified configuration

## Layout

- [cmd](/cmd) contains simulation commands
- [lib](/lib) contains simulation libraries
- [lib/simulator](/lib/simulator) links the simulation components
- [lib/config](/lib/config) defines and parses simulation configurations
- [lib/connector](/lib/connector) contains (ZMQ based) simulation connector module
- [lib/engine](/lib/engine) contains the core simulation engine
- [lib/medium](/lib/medium) contains the wireless medium emulation
- [lib/runner](/lib/runner) contains the client application runner
- [libyawns](/libyawns) contains the libyawns C library for client nodes as well as go bindings for testing these

## Licence

This project is licensed using the GNU Affero General Public License v3.0 (AGPL-3.0+), see [here](https://choosealicense.com/licenses/agpl-3.0/#) for a summary.

---

If you have any questions, comments, or suggestions, feel free to open an issue or a pull request.

