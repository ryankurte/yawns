# Open Network Simulator

A Next-Generation Wireless Sensor Network Simulation Engine for Wireless Sensor Network (WSN) Research and Development.

This is designed to assist the development and evaluation of WSN operating systems, protocols, and applications in a programmable virtual environment, with a focus on replicating the radio environment of a physical environment in a virtual space.

This project consists of two parts. The core simulator (ons) that manages clients and connections, simulates the medium and environment, and processes Events to the simulation environment, and libons that implements a C Language radio interface for use in the node application.

## Status

[![Build Status](https://travis-ci.com/ryankurte/ons.svg?token=s4CML2iJ2hd54vvqz5FP&branch=master)](https://travis-ci.com/ryankurte/ons)

Very early prototype

## Goals

- Common Medium interface
  - Definition of wireless medium as appropriate for simulation tasks
- Plugin support
  - Standard Plugin Interface
- PCAP streaming, file writing
- Management interface
- Platform / OS independent
- OpenGL / Map Visualisation

## Usage

ONS is designed to be platform and network agnostic. To simulate a given platform

1. Install ons
2. Create a wrapper for libons to adapt to the system under test
3. Create a simulation configuration file
4. Launch ons with the specified configuration

## Layout

- [lib](/lib) contains simulation libraries
- [lib/engine](/lib/simulator) links the simulation components
- [lib/engine](/lib/engine) contains the core simulation engine
- [lib/connector](/lib/connector) contains simulation connector module
- [lib/medium](/lib/medium) contains the wireless medium emulation
- [libons](/libons) contains the libons C library for client nodes as well as go binding for testing these

## Licence

TODO.