/**
 * OpenNetworkSim CZMQ/Protobuf Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#include "ons/protocol.h"

#include <stdint.h>

#include "protocol/ons.pb-c.h"


// ONS_DEBUG macro controls debug printing
#ifdef ONS_DEBUG
#include <stdio.h>
#define ONS_DEBUG_PRINT(...) printf(__VA_ARGS__)
#else
#define ONS_DEBUG_PRINT(...)
#endif



// Send an ONS message with the specified type
int ons_send_msg(struct ons_s *ons, uint8_t *data, uint16_t length)
{
    return zsock_send(ons->sock, "b", data, length);
}

int ons_send_pb(struct ons_s *ons, Base* message) {
    void* buf;

    size_t size = base__get_packed_size(message);

    buf = malloc(size + 1);

    size_t encoded_len = base__pack(message, buf);

    int res = ons_send_msg(ons, buf, encoded_len);

    free(buf);

    return res;
}

RFInfo ons_build_rfinfo(char* band, int channel) {
    RFInfo info = RFINFO__INIT;

    info.band = band;
    info.has_channel = 1;
    info.channel = channel;

    return info;
}

int ons_send_register(struct ons_s *ons, char* address) {
    Base base = BASE__INIT;
    Register reg = REGISTER__INIT;

    reg.address = address;

    base.message_case = BASE__MESSAGE_REGISTER;
    base.register_ = &reg;

    return ons_send_pb(ons, &base);
}

int ons_send_deregister(struct ons_s *ons, char* address) {
    Base base = BASE__INIT;
    Deregister dereg = DEREGISTER__INIT;

    dereg.address = address;

    return ons_send_pb(ons, &base);
}

int ons_send_packet(struct ons_s *ons, char* band, int32_t channel, uint8_t *data, uint16_t length) {
    Base base = BASE__INIT;
    Packet packet = PACKET__INIT;
    RFInfo info = RFINFO__INIT;

    packet.has_data = 1;
    packet.data.len = length;
    packet.data.data = data;

    info.band = band;
    info.has_channel = 1;
    info.channel = channel;
    packet.info = &info;

    base.message_case = BASE__MESSAGE_PACKET;
    base.packet = &packet;
   
    return ons_send_pb(ons, &base);
}

int ons_send_rssi_req(struct ons_s *ons, char* band, int channel) {
    Base base = BASE__INIT;
    RSSIReq req = RSSIREQ__INIT;

    RFInfo info = ons_build_rfinfo(band, channel);
    req.info = &info;

    base.message_case = BASE__MESSAGE_RSSI_REQ;
    base.rssireq = &req;

    return ons_send_pb(ons, &base);
}

int ons_send_start_receive(struct ons_s *ons, char* band, int channel) {
    Base base = BASE__INIT;
    StartReceive startreceive = START_RECEIVE__INIT;

    RFInfo info = ons_build_rfinfo(band, channel);
    startreceive.info = &info;

    base.message_case = BASE__MESSAGE_START_RECEIVE;
    base.startreceive = &startreceive;

    return ons_send_pb(ons, &base);
}

int ons_send_stop_receive(struct ons_s *ons, char* band, int channel) {
    Base base = BASE__INIT;
    StopReceive stopreceive = STOP_RECEIVE__INIT;

    RFInfo info = ons_build_rfinfo(band, channel);
    stopreceive.info = &info;

    base.message_case = BASE__MESSAGE_STOP_RECEIVE;
    base.stopreceive = &stopreceive;

    return ons_send_pb(ons, &base);
}

int ons_send_event(struct ons_s *ons, char* data) {
    Base base = BASE__INIT;
    Event event = EVENT__INIT;

    event.data = data;

    return ons_send_pb(ons, &base);
}
