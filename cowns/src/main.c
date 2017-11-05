/**
 * OpenNetworkSim CZMQ Radio Driver Example
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

#include <stdio.h>
#include <signal.h>

#include "owns/owns.h"
#include "owns/fifteenfour.h"

static volatile int running = 1;

void interrupt_handler(int x)
{
    running = 0;
}

void run_master(uint16_t addr, struct ons_radio_s *radio);
void run_slave(uint16_t addr, struct ons_radio_s *radio);

int main(int argc, char **argv)
{
    int res;

    printf("ONS Example Client\n");

    if (argc != 4) {
        printf("%s requires 4 arguments\n", argv[0]);
        return 0;
    }

    char *server_address = argv[1];
    char *local_address = argv[2];
    char *band = argv[3];

    uint16_t net_address = (uint16_t)strtol(local_address, NULL, 0);

    printf("Server Address: %s\n", server_address);
    printf("Local Address: %s (%d)\n", local_address, net_address);

    struct ons_s ons;
    struct ons_config_s config;
    res = ONS_init(&ons, server_address, local_address, &config);
    if (res < 0) {
        printf("Error %d creating ONS connector\n", res);
        return -1;
    }

    struct ons_radio_s radio;
    res = ONS_radio_init(&ons, &radio, band);
    if (res < 0) {
        printf("Error %d creating ONS virtual radio\n", res);
        return -1;
    }

    signal(SIGINT, interrupt_handler);

    usleep(10000);

    res = ONS_radio_start_receive(&radio, 0);
    if (res < 0) {
        printf("Error entering receive mode\n");
    }

    if (net_address == 0x01) {
        run_master(net_address, &radio);
    } else {
        run_slave(net_address, &radio);
    }

    printf("Exiting\n");

    ONS_radio_close(&ons, &radio);

    ONS_close(&ons);

    return 0;
}

uint16_t crc16_ccit_kermit(uint32_t len, uint8_t *data)
{
    uint16_t crc = 0x0000;

    for (uint32_t i = 0; i < len; i++) {
        uint8_t d = data[i];
        uint16_t q;

        q = (crc ^ d) & 0x0f;
        crc = (crc >> 4) ^ (q * 0x1081);
        q = (crc ^ (d >> 4)) & 0xf;
        crc = (crc >> 4) ^ (q * 0x1081);
    }

    return crc;
}

void run_master(uint16_t addr, struct ons_radio_s *radio)
{
    uint16_t seq = 0;
    int res;

    while (running) {
        struct fifteen_four_header_s header_out = FIFTEEN_FOUR_DEFAULT_HEADER(0x01, addr, addr + 1, seq);
        uint8_t test_data[] = {1, (addr & 0xFF), (addr >> 8)};

        uint8_t packet[sizeof(struct fifteen_four_header_s) + sizeof(test_data) + 2];
        memcpy(packet, &header_out, sizeof(struct fifteen_four_header_s));
        memcpy(packet + sizeof(struct fifteen_four_header_s), test_data, sizeof(test_data));

        uint16_t crc = crc16_ccit_kermit(sizeof(packet) - 2, packet);
        memcpy(packet + sizeof(struct fifteen_four_header_s) + sizeof(test_data), &crc, sizeof(crc));

        res = ONS_radio_send(radio, 0, packet, sizeof(packet));
        if (res < 0) {
            printf("ONS send error: %d\n", res);
        } else {
            while ((res = ONS_radio_check_send(radio)) == 0) {
                usleep(1000);
            }
        }
        ONS_radio_sleep();

        seq++;
        sleep(2);
    }
}

struct hops_s {
    uint8_t count;
    uint16_t addresses[16];
} __attribute__((packed));

void run_slave(uint16_t addr, struct ons_radio_s *radio)
{
    uint16_t seq = 0;
    uint8_t data[256];
    uint16_t len;
    int res;

    ONS_radio_start_receive(radio, 0);

    while (running) {
        res = ONS_radio_check_receive(radio);
        if (res > 0) {

            ONS_radio_get_received(radio, sizeof(data), data, &len);
            //ONS_print_arr("Received", data, len);

            struct fifteen_four_header_s *header_in = (struct fifteen_four_header_s *)&data[0];
            if (header_in->dest == addr) {
                header_in->seq = seq;
                header_in->src = addr;
                header_in->dest = addr + 1;

                #if 1
                struct hops_s * hops = (struct hops_s*) &data[sizeof(struct fifteen_four_header_s)];
                hops->addresses[hops->count] = addr;
                hops->count += 1;
                #endif
                len += 2;

                uint16_t crc = crc16_ccit_kermit(len - 2, data);
                memcpy(&data[len - 2], &crc, sizeof(crc));

                res = ONS_radio_send(radio, 0, data, len);
                if (res < 0) {
                    printf("ONS send error: %d\n", res);
                } else {
                    while ((res = ONS_radio_check_send(radio)) == 0) {
                        usleep(1000);
                    }
                }

                seq++;
            }

            //ONS_radio_start_receive(radio, 0);
        }

        usleep(10000);
    }
}
