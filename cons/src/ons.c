/**
 * OpenNetworkSim CZMQ/Protobuf Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#include "ons/ons.h"

#include <stdint.h>
#include <pthread.h>
#include <signal.h>
#include <czmq.h>
#include <semaphore.h>

#include "ons/protocol.h"
#include "protocol/ons.pb-c.h"


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

    pthread_mutex_init(&ons->rssi_mutex, NULL);
    pthread_mutex_init(&ons->rx_mutex, NULL);

    // Start listener thread
    ons->running = 1;
    pthread_create(&ons->thread, NULL, ons_handle_receive, ons);

    // Send message to register with server
    ons_send_register(ons, ons->local_address);

    return 0;
}

int ONS_send(struct ons_s *ons, uint8_t *data, uint16_t length)
{   
    return ons_send_packet(ons, data, length);
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

    pthread_mutex_lock(&ons->rx_mutex);

    if (ons->receive_length == 0) {
        pthread_mutex_unlock(&ons->rx_mutex);
        return 0;
    }

    *len = (ons->receive_length > max_len) ? max_len : ons->receive_length;
    memcpy(data, (const void *)ons->receive_data, *len);

    ons->receive_length = 0;

    pthread_mutex_unlock(&ons->rx_mutex);

    return 1;
}

int ONS_get_rssi(struct ons_s *ons, float* rssi)
{
    int res;

    ONS_DEBUG_PRINT("[ONCS] get rssi\n");

    ons->rssi_received = false;

    // TryLock in case mutex already locked
    pthread_mutex_trylock(&ons->rssi_mutex);

    // Send get CCA message
    ons_send_rssi_req(ons, "", 0);

    // Await cca mutex unlock from onsc thread
    res = pthread_mutex_lock(&ons->rssi_mutex);
    if (res < 0) {
        perror("[ONSC] rssi mutex lock error");
        return -1;
    }

    // Copy CCA
    *rssi = ons->rssi;
    bool rssi_received = ons->rssi_received;

    // Return mutex to unlocked state
    pthread_mutex_unlock(&ons->rssi_mutex);

    // Check a CCA message was received
    if (rssi_received != true) {
        ONS_DEBUG_PRINT("[ONCS] no rssi response received\n");
        return -2;
    }

    ONS_DEBUG_PRINT("[ONCS] got rssi value OK (%.2f)\n", *rssi);

    return 0;
}

int ONS_close(struct ons_s *ons)
{
    ONS_DEBUG_PRINT("[ONSC] Closing connector\n");

    ons->running = false;

    pthread_kill(ons->thread, SIGINT);

    pthread_join(ons->thread, NULL);

    pthread_mutex_destroy(&ons->rssi_mutex);
    pthread_mutex_destroy(&ons->rx_mutex);

    zsock_destroy(&ons->sock);

    ONS_DEBUG_PRINT("[ONSC] Closed\n");

    return 0;
}

void ONS_print_arr(char* name, uint8_t* data, uint16_t length)
{
    ONS_DEBUG_PRINT("%s (length: %d): ", name, length);
    for (int i = 0; i < length; i++) {
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

    ONS_DEBUG_PRINT("[ONSC THREAD] Starting recieve thread\n");

    // Bind exit handler to interrupt handler to avoid unhandled exits
    signal(SIGINT, exit_handler);

    while (ons->running) {

        res = zsock_recv(ons->sock, "b", &zdata, &zsize);
        if (res == 0) {

            ONS_print_arr("[ONSC THREAD] Received Data", zdata, zsize);

            Base *base = base__unpack(NULL, zsize, zdata);

            switch(base->message_case) {
                case BASE__MESSAGE_PACKET:

                // Check received packet is valid
                if((base->packet == NULL) || (base->packet->has_data == 0)) {
                    ONS_DEBUG_PRINT("[ONCS THREAD] invalid packet\n");
                    break;
                }

                // Copy data into local buffer
                int max_size = (base->packet->data.len > ONS_BUFFER_LENGTH) ? ONS_BUFFER_LENGTH - 1 : base->packet->data.len;
                pthread_mutex_lock(&ons->rx_mutex);
                memcpy((void *)ons->receive_data, base->packet->data.data, max_size);
                ons->receive_length = max_size;
                ONS_print_arr("[ONSC THREAD] Received packet", ons->receive_data, ons->receive_length);
                pthread_mutex_unlock(&ons->rx_mutex);

                break;

                case BASE__MESSAGE_RSSI_RESP:

                // Check RSSI packet is valid
                if((base->rssiresp == NULL) || (base->rssiresp->has_rssi == 0)) {
                    ONS_DEBUG_PRINT("[ONCS THREAD] invalid rssi response\n");
                    break;
                }

                // Copy RSSI data and signal receipt
                ons->rssi = base->rssiresp->rssi;
                ons->rssi_received = true;
                ONS_DEBUG_PRINT("[ONCS THREAD] got rssi response %.2f\n", ons->rssi);
                pthread_mutex_unlock(&ons->rssi_mutex);
                break;

                default:
                ONS_DEBUG_PRINT("[ONCS THREAD] unrecognised type %d\n", base->message_case);
            }

            base__free_unpacked(base,NULL);
            free(zdata);
        }
    }

    ONS_DEBUG_PRINT("[ONSC THREAD] Exiting recieve thread\n");

    return NULL;
}

