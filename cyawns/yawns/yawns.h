/**
 * OpenNetworkSim CZMQ Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

#ifndef YAWNS_H
#define YAWNS_H

#include <stdint.h>
#include <float.h>
#include <pthread.h>
#include <semaphore.h>

#include "zmq.h"
#include "czmq.h"

#ifdef __cplusplus
extern "C" {
#endif

#define ONS_MAX_RADIOS 16     //!< Maximum Virtual Radio Interfaces per ONS connector
#define ONS_STRING_LENGTH 64  //!< Maximum string length
#define ONS_BUFFER_LENGTH 256 //!< Maximum buffer length

// ONS radio state enumerations
enum ons_radio_state_e {
    ONS_RADIO_STATE_IDLE = 0,
    ONS_RADIO_STATE_RECEIVE = 1,
    ONS_RADIO_STATE_RECEIVING = 2,
    ONS_RADIO_STATE_TRANSMITTING = 3,
    ONS_RADIO_STATE_SLEEP = 4
};

// ONS event enumerations
enum ons_radio_event_e {
    ONS_RADIO_EVENT_NONE = 0,
    ONS_RADIO_EVENT_PACKET_RECEIVED = 1,
    ONS_RADIO_EVENT_SEND_DONE = 2,
};

// ONS connector configuration
struct ons_config_s {
    bool intercept_signals;
    bool debug_prints;
};

#define ONS_CONFIG_DEFAULT \
    {                      \
        true,              \
        false              \
    }

// ONS connector instance
struct ons_s {
    uint8_t local_address[ONS_STRING_LENGTH];
    uint8_t ons_address[ONS_STRING_LENGTH];

    zsock_t *sock;

    uint8_t running;
    pthread_t thread;

    pthread_mutex_t radios_mutex;
    struct ons_radio_s *radios[ONS_MAX_RADIOS];
    uint32_t radio_count;

    struct ons_config_s *config;
};

typedef void (*ons_radio_cb_f) (void* ctx, uint32_t event);

// ONS radio instance
struct ons_radio_s {
    struct ons_s *connector;
    char band[128];

    pthread_mutex_t rx_mutex;
    volatile uint16_t receive_length;
    volatile uint8_t receive_data[ONS_BUFFER_LENGTH];

    pthread_mutex_t tx_mutex;
    volatile bool tx_complete;

    pthread_mutex_t rssi_mutex;
    volatile float rssi;
    volatile float rssi_received;

    pthread_mutex_t state_mutex;
    volatile uint32_t state;
    volatile float state_received;

    ons_radio_cb_f cb;
    void* cb_ctx;
};

// Initialise the ONS connector
int ONS_init(struct ons_s *ons, char *ons_address, char *local_address, struct ons_config_s *config);

// Print the ONS connector status
int ONS_status(struct ons_s *ons);

// Send an ONS event
int YAWNS_event(struct ons_s *ons, uint8_t* name, uint8_t* data, size_t len);

// Create a for a specific band using the ons connector
int ONS_radio_init(struct ons_s *ons, struct ons_radio_s *radio, char *band);

// Attach an event callback to a radio
int ONS_radio_set_cb(struct ons_radio_s *radio, ons_radio_cb_f cb, void* ctx);

// Fetch the radio state
int ONS_radio_get_state(struct ons_radio_s *radio, uint32_t *state);

// Send a data packet using the connector
int ONS_radio_send(struct ons_radio_s *radio, int32_t channel, uint8_t *data, uint16_t length);

// Check for data packet send completion
int ONS_radio_check_send(struct ons_radio_s *radio);

// Cause a radio to enter receive mode
int ONS_radio_start_receive(struct ons_radio_s *radio, int32_t channel);

// Cause a radio to exit receive mode
int ONS_radio_stop_receive(struct ons_radio_s *radio);

// Check whether a packet has been received
int ONS_radio_check_receive(struct ons_radio_s *radio);

// Fetch a received packet
int ONS_radio_get_received(struct ons_radio_s *radio, uint16_t max_len, uint8_t *data, uint16_t *len);

// Put a radio into sleep mode
int ONS_radio_sleep(struct ons_radio_s *radio);

// Fetch rssi for a given band and channel
int ONS_radio_get_rssi(struct ons_radio_s *radio, int32_t channel, float *rssi);

// Close the ONS radio
int ONS_radio_close(struct ons_s *ons, struct ons_radio_s *radio);

// Set a field in the simulation
int ONS_set_field(struct ons_s *ons, char* name, char* data_str);

// Set a field in the simulation with formatted print
int ONS_set_fieldf(struct ons_s *ons, char* name, char* format, ...);

// Close the ONS connector
int ONS_close(struct ons_s *ons);

// Helper to print arrays
void ONS_print_arr(char *name, uint8_t *data, uint16_t len);

#ifdef __cplusplus
}
#endif

#endif
