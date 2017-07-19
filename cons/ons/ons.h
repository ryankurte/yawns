/**
 * OpenNetworkSim CZMQ Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#ifndef ONS_H
#define ONS_H

#include <stdint.h>
#include <float.h>
#include <pthread.h>
#include <semaphore.h>

#include "czmq.h"

#define ONS_MAX_RADIOS      16
#define ONS_STRING_LENGTH   64      // Maximum string length
#define ONS_BUFFER_LENGTH   256     // Maximum buffer length


// ONS connector instance
struct ons_s {
    uint8_t local_address[ONS_STRING_LENGTH];
    uint8_t ons_address[ONS_STRING_LENGTH];

    zsock_t* sock;

    uint8_t running;
    pthread_t thread;
    
    pthread_mutex_t radios_mutex;
    struct ons_radio_s *radios[ONS_MAX_RADIOS];
};

// ONS radio instance
struct ons_radio_s {
    struct ons_s* connector;
    char band[128];

    pthread_mutex_t rx_mutex;
    volatile uint16_t receive_length;
    volatile uint8_t receive_data[ONS_BUFFER_LENGTH];

    pthread_mutex_t tx_mutex;
    volatile bool tx_complete;

    pthread_mutex_t rssi_mutex;
    volatile float rssi;
    volatile float rssi_received;
};

// Initialise the ONS connector
int ONS_init(struct ons_s *ons, char* ons_address, char* local_address);

// Create a for a specific band using the ons connector
int ONS_radio_init(struct ons_s *ons, struct ons_radio_s *radio, char* band);

// Send a data packet using the connector
int ONS_radio_send(struct ons_radio_s *radio, int32_t channel, uint8_t *data, uint16_t length);

// Check for data packet send completion
int ONS_radio_check_send(struct ons_radio_s *radio);

// Cause a radio to enter receive mode
int ONS_radio_start_receive(struct ons_radio_s *radio, int32_t channel);

// Cause a radio to exit receive mode
int ONS_radio_stop_receive(struct ons_radio_s *radio) ;

// Check whether a packet has been received
int ONS_radio_check_receive(struct ons_radio_s *radio);

// Fetch a received packet 
int ONS_radio_get_received(struct ons_radio_s *radio, uint16_t max_len, uint8_t* data, uint16_t* len);

// Fetch rssi for a given band and channel
int ONS_radio_get_rssi(struct ons_radio_s *radio, int32_t channel, float* rssi);

// Close the ONS radio
int ONS_radio_close(struct ons_s *ons, struct ons_radio_s *radio);

// Close the ONS connector
int ONS_close(struct ons_s *ons);

// Helper to print arrays
void ONS_print_arr(char* name, uint8_t* data, uint16_t len);

#endif

