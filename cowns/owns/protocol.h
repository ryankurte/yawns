/**
 * OpenNetworkSim CZMQ Radio Driver Library
 * Communication protocol helpers
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#ifndef ONS_PROTOCOL_H
#define ONS_PROTOCOL_H

#include <stdint.h>

#include "owns/owns.h"

#ifdef __cplusplus
extern "C" {
#endif

int ons_send_register(struct ons_s *ons, char* address);
int ons_send_deregister(struct ons_s *ons, char* address);
int ons_send_packet(struct ons_s *ons, char* band, int32_t channel, uint8_t *data, uint16_t length);
int ons_send_rssi_req(struct ons_s *ons, char* band, int channel);
int ons_send_state_req(struct ons_s *ons, char* band);
int ons_send_start_receive(struct ons_s *ons, char* band, int channel);
int ons_send_idle(struct ons_s *ons, char* band);
int ons_send_sleep(struct ons_s *ons, char* band);
int ons_send_event(struct ons_s *ons, char* data);
int ons_send_field_set(struct ons_s *ons, char* name, uint8_t* data, size_t len);
int ons_send_field_req(struct ons_s *ons, char* name);    

#ifdef __cplusplus
}
#endif

#endif
