/**
 * OpenNetworkSim CZMQ Radio Driver Example
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */
 
#include <stdio.h>
#include <signal.h>

#include "ons/ons.h"

static volatile int running = 1;

void interrupt_handler(int x) {
    running = 0;
}

int main(int argc, char** argv) {
    int res;

    printf("ONS Example Client\n");

    if (argc != 3) {
        printf("%s requires 3 arguments\n", argv[0]);
        return 0;
    }

    char* server_address = argv[1];
    char* local_address = argv[2];

    printf("Server Address: %s\n", server_address);
    printf("Local Address: %s\n", local_address);

    struct ons_s ons;
    res = ONS_init(&ons, server_address, local_address);
    if (res < 0) {
        printf("Error %d creating ONS connector\n", res);
        return -1;
    }

    struct ons_radio_s radio;
    res = ONS_radio_init(&ons, &radio, "ISM-433MHz");
    if (res < 0) {
        printf("Error %d creating ONS connector\n", res);
        return -1;
    }

    uint8_t data[256];
    uint16_t len;

    signal(SIGINT, interrupt_handler);

    int count = 0;

    while(running) {
        int res = ONS_radio_check_receive(&radio);
        if (res > 0) {
            ONS_radio_get_received(&radio, sizeof(data), data, &len);
            ONS_print_arr("Received", data, len);
            ONS_radio_send(&radio, 0, data, len);
        }

        data[0] = count ++;

        res = ONS_radio_send(&radio, 0, data, 1);
        if (res < 0) {
            printf("ONS send error: %d\n", res);
        }

        sleep(30);
    }

    printf("Exiting\n");

    ONS_radio_close(&ons, &radio);

    ONS_close(&ons);

    return 0;
}

