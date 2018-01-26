/**
 * OpenNetworkSim CZMQ/Protobuf Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

#include "owns/owns.h"

#include <stdint.h>
#include <stdbool.h>
#include <assert.h>
#include <pthread.h>
#include <semaphore.h>

#include "zmq.h"
#include "czmq.h"
#include "owns/protocol.h"
#include "ons.pb-c.h"

#define ONS_DEBUG
#define ONS_INFO

// ONS_DEBUG macro controls debug printing
#ifndef ONS_PRINTF
#include <stdio.h>
#define ONS_PRINTF(...) printf(__VA_ARGS__)
#endif

#define ONS_CORE_PRINT(...) if (ons->config->debug_prints) ONS_PRINTF(__VA_ARGS__);
#define ONS_RADIO_PRINT(...) if (radio->connector->config->debug_prints) ONS_PRINTF(__VA_ARGS__);

/** Internal Function Prototypes **/

void *ons_handle_receive(void *ctx);
int ons_send_msg(struct ons_s *ons, uint32_t type, uint8_t *data, uint16_t length);

/** External Functions **/

int ONS_init(struct ons_s *ons, char *ons_address, char *local_address, struct ons_config_s *config)
{
    // Copy configuration
    strncpy((char *)ons->ons_address, ons_address, ONS_STRING_LENGTH);
    strncpy((char *)ons->local_address, local_address, ONS_STRING_LENGTH);
    ons->radio_count = 0;
    ons->config = config;

    ONS_CORE_PRINT("[ONSC] Connecting to '%s' as '%s'\n", ons_address, local_address);

    // Create ZMQ socket
    ons->sock = zsock_new_dealer(ons_address);

    // Initialise radio list
    pthread_mutex_init(&ons->radios_mutex, NULL);
    for (int i = 0; i < ONS_MAX_RADIOS; i++) {
        ons->radios[i] = NULL;
    }

    // Start listener thread
    ons->running = 1;
    pthread_create(&ons->thread, NULL, ons_handle_receive, ons);

    // Send message to register with server
    return ons_send_register(ons, (char *)ons->local_address);
}

int ONS_status(struct ons_s *ons)
{
    printf("OWNS connector status: ");

    if (ons->running != 0) {
        printf("Running\n");
        printf("\t- Socket Type: %s\n", zsock_type_str(ons->sock));
    } else {
        printf("Not connected\n");
    }

    return 0;
}

int ONS_radio_init(struct ons_s *ons, struct ons_radio_s *radio, char *band)
{
    radio->connector = NULL;
    radio->cb = NULL;
    strncpy(radio->band, band, sizeof(radio->band) - 1);

    // Init mutexes
    pthread_mutex_init(&radio->rssi_mutex, NULL);
    pthread_mutex_init(&radio->state_mutex, NULL);
    pthread_mutex_init(&radio->rx_mutex, NULL);
    pthread_mutex_init(&radio->tx_mutex, NULL);

    // Attach radio instance to connector
    pthread_mutex_lock(&ons->radios_mutex);
    for (int i = 0; i < ONS_MAX_RADIOS; i++) {
        if (ons->radios[i] == NULL) {
            ons->radios[i] = radio;
            radio->connector = ons;
            ons->radio_count++;
            break;
        }
    }
    pthread_mutex_unlock(&ons->radios_mutex);

    if (radio->connector == NULL) {
        return -1;
    }

    return 0;
}

int ONS_radio_close(struct ons_s *ons, struct ons_radio_s *radio)
{
    ONS_CORE_PRINT("[ONSC] Closing radio\n");

    // Remove from radio list
    pthread_mutex_lock(&ons->radios_mutex);
    for (int i = 0; i < ONS_MAX_RADIOS; i++) {
        if (ons->radios[i] == radio) {
            ons->radios[i] = NULL;
            ons->radio_count--;
            break;
        }
    }
    pthread_mutex_unlock(&ons->radios_mutex);

    // Remove radio mutexes
    pthread_mutex_destroy(&radio->rssi_mutex);
    pthread_mutex_destroy(&radio->rx_mutex);

    return 0;
}

int ONS_close(struct ons_s *ons)
{
    ONS_CORE_PRINT("[ONSC] Closing connector\n");

    ons->running = false;

    pthread_kill(ons->thread, SIGINT);

    pthread_join(ons->thread, NULL);

    zsock_destroy(&ons->sock);

    ONS_CORE_PRINT("[ONSC] Closed\n");

    return 0;
}

int ONS_radio_set_cb(struct ons_radio_s *radio, ons_radio_cb_f cb, void* ctx)
{
    if (radio == NULL) {
        return -1;
    }

    radio->cb = cb;
    radio->cb_ctx = ctx;
    return 0;
}

int ONS_radio_send(struct ons_radio_s *radio, int32_t channel, uint8_t *data, uint16_t length)
{
    if (radio == NULL) {
        return -1;
    }

    ONS_RADIO_PRINT("[ONCS] send %d bytes on channel %d\n", length, channel);

    radio->tx_complete = false;
    return ons_send_packet(radio->connector, radio->band, channel, data, length);
}

int ONS_radio_check_send(struct ons_radio_s *radio)
{
    if (radio == NULL) {
        return -1;
    }

    ONS_RADIO_PRINT("[ONCS] check send\n");

    if (radio->tx_complete) {
        return 1;
    } else {
        return 0;
    }
}

int ONS_radio_start_receive(struct ons_radio_s *radio, int32_t channel)
{
    if (radio == NULL) {
        return -1;
    }

    ONS_RADIO_PRINT("[ONCS] start receive\n");

    return ons_send_start_receive(radio->connector, radio->band, channel);
}

int ONS_radio_stop_receive(struct ons_radio_s *radio)
{
    if (radio == NULL) {
        return -1;
    }

    ONS_RADIO_PRINT("[ONCS] stop receive\n");

    return ons_send_idle(radio->connector, radio->band);
}

int ONS_radio_sleep(struct ons_radio_s *radio)
{
    if (radio == NULL) {
        return -1;
    }

    ONS_RADIO_PRINT("[ONCS] radio sleep\n");

    return ons_send_sleep(radio->connector, radio->band);
}


int ONS_radio_check_receive(struct ons_radio_s *radio)
{
    if (radio == NULL) {
        return -1;
    }

    ONS_RADIO_PRINT("[ONCS] check receive\n");

    if (radio->receive_length > 0) {
        return 1;
    }
    return 0;
}

int ONS_radio_get_received(struct ons_radio_s *radio, uint16_t max_len, uint8_t *data, uint16_t *len)
{
    if (radio == NULL) {
        return -1;
    }

    ONS_RADIO_PRINT("[ONCS] get received\n");

    pthread_mutex_lock(&radio->rx_mutex);

    if (radio->receive_length == 0) {
        pthread_mutex_unlock(&radio->rx_mutex);
        return 0;
    }

    *len = (radio->receive_length > max_len) ? max_len : radio->receive_length;
    memcpy(data, (const void *)radio->receive_data, *len);

    radio->receive_length = 0;

    pthread_mutex_unlock(&radio->rx_mutex);

    return 1;
}

int ONS_radio_get_state(struct ons_radio_s *radio, uint32_t *state)
{
    if (radio == NULL) {
        return -1;
    }

    int res;

    ONS_RADIO_PRINT("[ONCS] get state\n");

    radio->state_received = false;

    // TryLock in case mutex already locked
    pthread_mutex_trylock(&radio->state_mutex);

    // Send get CCA message
    ons_send_state_req(radio->connector, radio->band);

    // Await cca mutex unlock from onsc thread
    res = pthread_mutex_lock(&radio->state_mutex);
    if (res < 0) {
        perror("[ONSC] rssi mutex lock error");
        return -1;
    }

    // Copy CCA
    switch (radio->state) {
    case RFSTATE__IDLE:
        *state = ONS_RADIO_STATE_IDLE;
        break;
    case RFSTATE__RECEIVE:
        *state = ONS_RADIO_STATE_RECEIVE;
        break;
    case RFSTATE__RECEIVING:
        *state = ONS_RADIO_STATE_RECEIVING;
        break;
    case RFSTATE__TRANSMITTING:
        *state = ONS_RADIO_STATE_TRANSMITTING;
        break;
    }

    *state = radio->state;
    bool state_received = radio->state_received;

    // Return mutex to unlocked state
    pthread_mutex_unlock(&radio->state_mutex);

    // Check a CCA message was received
    if (state_received != true) {
        ONS_RADIO_PRINT("[ONCS] no state response received\n");
        return -2;
    }

    ONS_RADIO_PRINT("[ONCS] got state value OK (%d)\n", *state);

    return 0;
}

int ONS_radio_get_rssi(struct ons_radio_s *radio, int32_t channel, float *rssi)
{
    if (radio == NULL) {
        return -1;
    }

    int res;

    ONS_RADIO_PRINT("[ONCS] get rssi\n");

    radio->rssi_received = false;

    // TryLock in case mutex already locked
    pthread_mutex_trylock(&radio->rssi_mutex);

    // Send get CCA message
    ons_send_rssi_req(radio->connector, radio->band, channel);

    // Await cca mutex unlock from onsc thread
    res = pthread_mutex_lock(&radio->rssi_mutex);
    if (res < 0) {
        perror("[ONSC] rssi mutex lock error");
        return -1;
    }

    // Copy CCA
    *rssi = radio->rssi;
    bool rssi_received = radio->rssi_received;

    // Return mutex to unlocked state
    pthread_mutex_unlock(&radio->rssi_mutex);

    // Check a CCA message was received
    if (rssi_received != true) {
        ONS_RADIO_PRINT("[ONCS] no rssi response received\n");
        return -2;
    }

    ONS_RADIO_PRINT("[ONCS] got rssi value OK (%.2f)\n", *rssi);

    return 0;
}

int ONS_set_field(struct ons_s *ons, char* name, char* data_str)
{
    return ons_send_field_set(ons, name, data_str);
}

int ONS_set_fieldf(struct ons_s *ons, char* name, char* format, ...)
{
    va_list args;
    char buff[1024];

    va_start(args, format);
    int n = vsnprintf(buff, sizeof(buff)-1, format, args);
    va_end(args);
    buff[n] = '\0';

    ONS_PRINTF("set field '%s' data: '%s'\n", name, buff);

    return ONS_set_field(ons, name, buff);
}

void ONS_print_arr(char *name, uint8_t *data, uint16_t length)
{
    ONS_PRINTF("%s (length: %d): ", name, length);
    for (int i = 0; i < length; i++) {
        ONS_PRINTF("%.2x ", data[i]);
    }
    ONS_PRINTF("\n");
}


/** Internal Functions **/

// Stub exit handler for signal binding
void exit_handler(int x) {}

// Find a radio instance by band
struct ons_radio_s *ons_get_radio(struct ons_s *ons, char *band)
{
    for (int i = 0; i < ONS_MAX_RADIOS; i++) {
        if (strcmp(band, ons->radios[i]->band) == 0) {
            return ons->radios[i];
        }
    }
    return NULL;
}

// ONS receiver thread
// The thread can be exited by setting ons->running = 0 then passing a SIGINT
void *ons_handle_receive(void *ctx)
{
    struct ons_s *ons = (struct ons_s *)ctx;
    uint8_t *zdata;
    size_t zsize;
    int res;

    ONS_CORE_PRINT("[ONSC THREAD] Starting receive thread\n");

    // Bind exit handler to interrupt handler to avoid unhandled exits
    if (ons->config->intercept_signals)
        signal(SIGINT, exit_handler);

    while (ons->running) {

        res = zsock_recv(ons->sock, "b", &zdata, &zsize);
        if (res == 0) {

            Base *base = base__unpack(NULL, zsize, zdata);
            struct ons_radio_s *radio = NULL;

            if (base == NULL) {
                ONS_CORE_PRINT("[ONSC THREAD] Error parsing message");
                continue;
            }

            pthread_mutex_lock(&ons->radios_mutex);

            switch (base->message_case) {
            case BASE__MESSAGE_PACKET:

                // Check received packet is valid
                if ((base->packet == NULL) || (base->packet->data.len == 0) || (base->packet->info == NULL)) {
                    ONS_CORE_PRINT("[ONCS THREAD] invalid packet\n");
                    break;
                }

                // Find matching radio instance
                radio = ons_get_radio(ons, base->packet->info->band);
                if (radio == NULL) {
                    ONS_CORE_PRINT("[ONCS THREAD] no radio found matching packet\n");
                    break;
                }

                // Copy data into local buffer
                int max_size = (base->packet->data.len > ONS_BUFFER_LENGTH) ? ONS_BUFFER_LENGTH - 1 : base->packet->data.len;
                pthread_mutex_lock(&radio->rx_mutex);
                memcpy((void *)radio->receive_data, base->packet->data.data, max_size);
                radio->receive_length = max_size;
                if (ons->config->debug_prints)
                    ONS_print_arr("[ONSC THREAD] Received packet", (uint8_t *)radio->receive_data, radio->receive_length);
                pthread_mutex_unlock(&radio->rx_mutex);

                if (radio->cb != NULL) {
                    radio->cb(radio->cb_ctx, ONS_RADIO_EVENT_PACKET_RECEIVED);
                }

                break;

            case BASE__MESSAGE_RSSI_RESP:
                // Check RSSI packet is valid
                if ((base == NULL) || (base->rssiresp == NULL) ||
                    (base->rssiresp->info == NULL) || (base->rssiresp->info->band == NULL)) {
                    ONS_CORE_PRINT("[ONCS THREAD] invalid rssi response (missing elements) %p %p %p\n",
                                   base->rssiresp, base->rssiresp->info, base->rssiresp->info->band
                                  );
                    break;
                }

                // Find matching radio instance
                radio = ons_get_radio(ons, base->rssiresp->info->band);
                if (radio == NULL) {
                    ONS_CORE_PRINT("[ONCS THREAD] no radio found matching rssi response\n");
                    break;
                }

                // Copy RSSI data and signal receipt
                radio->rssi = base->rssiresp->rssi;
                radio->rssi_received = true;
                ONS_CORE_PRINT("[ONCS THREAD] got rssi response %.2f\n", radio->rssi);
                pthread_mutex_unlock(&radio->rssi_mutex);
                break;

            case BASE__MESSAGE_STATE_RESP:
                // Check RSSI packet is valid
                if ((base == NULL) || (base->stateresp == NULL) ||
                    (base->stateresp->info == NULL) || (base->stateresp->info->band == NULL)) {
                    ONS_CORE_PRINT("[ONCS THREAD] invalid state response (missing elements)\n");
                    break;
                }

                // Find matching radio instance
                radio = ons_get_radio(ons, base->stateresp->info->band);
                if (radio == NULL) {
                    ONS_CORE_PRINT("[ONCS THREAD] no radio found matching state response\n");
                    break;
                }

                // Copy RSSI data and signal receipt
                radio->state = base->stateresp->state;
                radio->state_received = true;
                ONS_CORE_PRINT("[ONCS THREAD] got state response %d\n", radio->state);
                pthread_mutex_unlock(&radio->state_mutex);
                break;

            case BASE__MESSAGE_SEND_COMPLETE:
                if (base == NULL || base->sendcomplete == NULL || base->sendcomplete->info == NULL || base->sendcomplete->info->band == NULL) {
                    ONS_CORE_PRINT("[ONCS THREAD] invalid send complete\n");
                    break;
                }

                // Find matching radio instance
                radio = ons_get_radio(ons, base->sendcomplete->info->band);
                if (radio == NULL) {
                    ONS_CORE_PRINT("[ONCS THREAD] no radio found matching rssi response\n");
                    break;
                }

                radio->tx_complete = true;

                if (radio->cb != NULL) {
                    radio->cb(radio->cb_ctx, ONS_RADIO_EVENT_SEND_DONE);
                }


                ONS_CORE_PRINT("[ONCS THREAD] got tx complete\n");
                break;

            default:
                ONS_CORE_PRINT("[ONCS THREAD] unrecognised type %d\n", base->message_case);
                if (ons->config->debug_prints)
                    ONS_print_arr("[ONSC THREAD] unrecognised data", zdata, zsize);
            }

            pthread_mutex_unlock(&ons->radios_mutex);

            base__free_unpacked(base, NULL);
            free(zdata);
        }
    }

    ONS_CORE_PRINT("[ONSC THREAD] Exiting receive thread\n");

    return NULL;
}
