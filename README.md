# Open Network Simulator

A Next-Generation Wireless Sensor Network Simulation Engine for Wireless Sensor Network (WSN) Research and Development.

This is designed to assist the development and evaluation of WSN operating systems, protocols, and applications in a programmable virtual environment, with a focus on replicating the radio environment of a physical environment in a virtual space.

This project consists of two parts. The core simulator (ons) that manages clients and connections, simulates the medium and environment, and processes Events to the simulation environment, and libons that implements a C Language radio interface for use in the node application.

## Status

[![Documentation](https://img.shields.io/badge/docs-godoc-blue.svg)](https://godoc.org/github.com/ryankurte/ons)
[![GitHub tag](https://img.shields.io/github/tag/ryankurte/ons.svg)](https://github.com/ryankurte/ons)
[![Build Status](https://travis-ci.com/ryankurte/ons.svg?token=s4CML2iJ2hd54vvqz5FP&branch=master)](https://travis-ci.com/ryankurte/ons)

Very early prototype

## Goals

- [ ] Event Engine
- [ ] Common Medium interface
  - [ ] Definition of wireless medium as appropriate for simulation tasks
- [ ] Plugin support
  - [X]Standard Plugin Interface
- [ ] PCAP streaming, file writing
- [ ] Runnable / Client Management
- [ ] Platform / OS independent
- [ ] OpenGL / Map Visualisation

## Dependencies

- sodium
- czmq
- go

### OSX

OSX dependencies can be installed with `brew install libsodium czmq go`, though you may prefer to use the more up-to-date official go package from [golang.org/](https://golang.org/dl/).



## Usage

ONS is designed to be platform and network agnostic. To simulate a given platform

1. Install ons
2. Create a wrapper for libons to adapt to the system under test
3. Create a simulation configuration file
4. Launch ons with the specified configuration

## Layout

- [lib](/lib) contains simulation libraries
- [lib/simulator](/lib/simulator) links the simulation components
- [lib/config](/lib/config) defines and parses simulation configurations
- [lib/connector](/lib/connector) contains (ZMQ based) simulation connector module
- [lib/engine](/lib/engine) contains the core simulation engine
- [lib/medium](/lib/medium) contains the wireless medium emulation
- [lib/runner](/lib/runner) contains the client application runner
- [libons](/libons) contains the libons C library for client nodes as well as go bindings for testing these

## Licence

TODO.

