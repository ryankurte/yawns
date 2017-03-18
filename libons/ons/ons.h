/**
 * OpenNetworkSim CZMQ Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#ifndef ONS_H
#define ONS_H

#include <stdint.h>
#include <pthread.h>
#include <semaphore.h>

#include "czmq.h"

#define ONS_STRING_LENGTH   64      // Maximum string length
#define ONS_BUFFER_LENGTH   256     // Maximum buffer length


struct ons_s {
    uint8_t local_address[ONS_STRING_LENGTH];
    uint8_t ons_address[ONS_STRING_LENGTH];

    zsock_t* sock;

    uint8_t running;
    pthread_t thread;
    
    pthread_mutex_t rx_mutex;
    volatile uint16_t receive_length;
    volatile uint8_t receive_data[ONS_BUFFER_LENGTH];

    pthread_mutex_t cca_mutex;
    volatile bool cca;
    volatile bool cca_received;
};

// Initialise the ONS connector
int ONS_init(struct ons_s *ons, char* ons_address, char* local_address);

// Send a data packet using the connector
int ONS_send(struct ons_s *ons, uint8_t *data, uint16_t length);

// Check for data packet send completion
int ONS_check_send(struct ons_s *ons);

// Check whether a packet has been received
int ONS_check_receive(struct ons_s *ons);

// Fetch a received packet
int ONS_get_received(struct ons_s *ons, uint16_t max_len, uint8_t* data, uint16_t* len);

// Fetch clear channel acknowledgement
int ONS_get_cca(struct ons_s *ons);

// Close the ONS connector
int ONS_close(struct ons_s *ons);

// Helper to print arrays
void ONS_print_arr(char* name, uint8_t* data, uint16_t len);

#endif

