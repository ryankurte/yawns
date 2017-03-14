/**
 * OpenNetworkSim CZMQ Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#include "ons/ons.h"

#include <stdint.h>
#include <pthread.h>
#include <signal.h>
#include <czmq.h>
#include <semaphore.h>

// ONS message enum
// Types must match go connector
enum ons_msg_e {
    ONS_MSG_REGISTER = 1,
    ONS_MSG_PACKET = 2,
    ONS_MSG_CCA_REQ = 3,
    ONS_MSG_CCA_RESP = 4
};


#define ONS_DEBUG

// ONS_DEBUG macro controls debug printing
#ifdef ONS_DEBUG
#include <stdio.h>
#define ONS_DEBUG_PRINT(...) printf(__VA_ARGS__)
#else
#define ONS_DEBUG_PRINT(...)
#endif


/** Internal Function Prototypes **/

void *ons_handle_receive(void* ctx);
int ons_send_msg(struct ons_s *ons, uint32_t type, uint8_t *data, uint16_t length);


/** External Functions **/

int ONS_init(struct ons_s *ons, char* ons_address, char* local_address)
{    
    // Copy configuration
    strncpy((char*)ons->ons_address, ons_address, ONS_STRING_LENGTH);
    strncpy((char*)ons->local_address, local_address, ONS_STRING_LENGTH);   
    
    ONS_DEBUG_PRINT("[ONSC] Init\n");
    ONS_DEBUG_PRINT("[ONSC] Connecting to '%s' as '%s'\n", ons_address, local_address);

    // Create ZMQ socket
    ons->sock = zsock_new_dealer(ons_address);

    // Create semaphores
    pthread_mutex_init(&ons->cca_mutex, NULL);
    pthread_cond_init(&ons->cca_cond, NULL);

    // Start listener thread
    ons->running = 1; 
    pthread_create(&ons->thread, NULL, ons_handle_receive, ons);

    // Send message to register with server
    ons_send_msg(ons, ONS_MSG_REGISTER, ons->local_address, strlen((char*)ons->local_address));

    return 0;
}


int ONS_send(struct ons_s *ons, uint8_t *data, uint16_t length) 
{
    ONS_print_arr("[ONSC] Send: ", data, length);
    return ons_send_msg(ons, ONS_MSG_PACKET, data, length);
}

int ONS_check_send(struct ons_s *ons)
{
    return 1;
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

int ONS_get_cca(struct ons_s *ons)
{
    int res;
    struct timespec ts;

    ONS_DEBUG_PRINT("[ONCS] get cca\n");

    // Reset mutex
    pthread_mutex_lock(&ons->cca_mutex);
    
    // Send get CCA message
    ons_send_msg(ons, ONS_MSG_CCA_REQ, NULL, 0);

    // Await cca mutex
    ts.tv_sec = time(NULL) + 2;
    res = pthread_cond_timedwait(&ons->cca_cond, &ons->cca_mutex, &ts);

    if ((res == -1) && (errno == ETIMEDOUT))  {
        ONS_DEBUG_PRINT("[ONSC] get_cca timeout");
        return -1;
    } else if (res == -1) {
        ONS_DEBUG_PRINT("[ONSC] get_cca error %d", errno);
        return -2;
    }

    ONS_DEBUG_PRINT("[ONCS] got cca value %d\n", ons->cca);
    
    return ons->cca;
}

int ONS_close(struct ons_s *ons)
{
    ONS_DEBUG_PRINT("[ONSC] Closing connector\n");

    ons->running = false;

    pthread_kill(ons->thread, SIGINT);

    pthread_join(ons->thread, NULL);

    pthread_mutex_destroy(&ons->cca_mutex);

    zsock_destroy(&ons->sock);

    ONS_DEBUG_PRINT("[ONSC] Closed\n");

    return 0;
}

void ONS_print_arr(char* name, uint8_t* data, uint16_t length) {
    ONS_DEBUG_PRINT("%s (length: %d): ", name, length);
    for(int i=0; i<length; i++) {
        ONS_DEBUG_PRINT("%.2x ", data[i]);
    }
    ONS_DEBUG_PRINT("\n");
}


/** Internal Functions **/

// Stub exit handler for signal binding
void exit_handler(int x) {}

// ONS receiver thread
// The thread can be exited by setting ons->running = 0 then passing a SIGINT
void *ons_handle_receive(void* ctx)
{
    struct ons_s *ons = (struct ons_s*) ctx;
    uint8_t *zdata;
    size_t zsize;
    uint8_t type = 8;
    int res;

    ONS_DEBUG_PRINT("[ONSC] Starting recieve thread\n");

    // Bind exit handler to interrupt handler to avoid unhandled exits
    signal(SIGINT, exit_handler);

    while(ons->running) {

        res = zsock_recv(ons->sock, "1b", &type, &zdata, &zsize);   
        if (res == 0) {
            
            ONS_DEBUG_PRINT("[ONCS] Received message type %u\n", type);
            ONS_print_arr("[ONSC] Data", zdata, zsize);

            int max_size = (zsize > ONS_BUFFER_LENGTH) ? ONS_BUFFER_LENGTH - 1: zsize;

            if (type == ONS_MSG_PACKET) {
                memcpy(ons->receive_data, zdata, max_size);
                ons->receive_length = max_size;

                ONS_print_arr("[ONSC] Received packet", ons->receive_data, ons->receive_length);

            } else if (type == ONS_MSG_CCA_RESP) {
                if (zsize != 1) {
                    ONS_DEBUG_PRINT("[ONCS] cca_resp invalid length\n");
                    break;
                }

                ons->cca = zdata[0];
                ONS_DEBUG_PRINT("[ONCS] got cca response %u\n", ons->cca);
                pthread_cond_signal(&ons->cca_cond);
                pthread_mutex_unlock(&ons->cca_mutex);

            } else {
                ONS_DEBUG_PRINT("[ONCS] unrecognised type\n");
            }
            
            free(zdata);
        }
    }

    ONS_DEBUG_PRINT("[ONSC] Exiting recieve thread\n");

    return NULL;
}

// Send an ONS message with the specified type
int ons_send_msg(struct ons_s *ons, uint32_t type, uint8_t *data, uint16_t length)
{
    return zsock_send(ons->sock, "1b", type, data, length);
}
