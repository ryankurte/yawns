/**
 * OpenNetworkSim CZMQ Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#include "ons/ons.h"

#include <stdint.h>
#include <pthread.h>

#include <czmq.h>


void *ons_handle_receive(void* ctx);


int ONS_init(struct ons_s *ons, char* ons_address, char* local_address)
{    
    // Copy configuration
    strncpy(ons->ons_address, ons_address, ONS_STRING_LENGTH);
    strncpy(ons->local_address, local_address, ONS_STRING_LENGTH);   
    
    // Create ZMQ socket
    ons->sock = zsock_new_dealer(ons_address);

    // Start listener thread
    ons->running = 1; 
    //pthread_create(&ons->thread, NULL, ons_handle_receive, ons);

    zstr_send(ons->sock, local_address);

    // Send message to register with server
    //ONS_send(ons, NULL, 0);

    return 0;
}

int ONS_send(struct ons_s *ons, uint8_t *data, uint16_t length)
{
    // Send
    int res = zsock_send(ons->sock, "bb", 
                            ons->local_address, sizeof(ons->local_address),
                            data, length);

    return res;
}

void *ons_handle_receive(void* ctx)
{
    struct ons_s *ons = (struct ons_s*) ctx;
    uint8_t *zdata;
    size_t zsize;

    while(ons->running) {

        int res = zsock_recv(ons->sock, "b", &zdata, &zsize);   
        if (res == 0) {
            int max_size = (zsize > ONS_BUFFER_LENGTH) ? ONS_BUFFER_LENGTH - 1: zsize;

            memcpy(ons->receive_data, zdata, max_size);
            ons->receive_length = max_size;
            
            free(zdata);
        }
    }

    return NULL;
}

int ONS_check_receive(struct ons_s *ons)
{
    if (ons->receive_length > 0) {
        return 1;
    }
    return 0;
}

int ONS_get_received(struct ons_s *ons, uint16_t max_len, uint8_t* data, uint16_t* len)
{
    if (ons->receive_length == 0) {
        return 0;
    }

    int max_size = (ons->receive_length > max_len) ? max_len : ons->receive_length;

    memcpy(data, ons->receive_data, max_size);

    *len = max_size;
 
    ons->receive_length = 0;
   
    return 1;
}

int ONS_close(struct ons_s *ons)
{

    ons->running = false;

    //pthread_join(ons->thread, NULL);

    zsock_destroy(&ons->sock);

    return 0;
}

