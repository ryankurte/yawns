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
    ONS_init(&ons, server_address, local_address);

    uint8_t data[256];
    uint16_t len;

    signal(SIGINT, interrupt_handler);

    while(running) {
        int res = ONS_check_receive(&ons);
        if (res > 0) {
            ONS_get_received(&ons, sizeof(data), data, &len);
            ONS_print_arr("Received", data, len);
            ONS_send(&ons, data, len);
        }
    }

    printf("Exiting\n");

    ONS_close(&ons);

    return 0;
}

