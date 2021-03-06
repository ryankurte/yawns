/**
 * OpenNetworkSim CZMQ/Protobuf Radio Driver Library
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */


#include "yawns/protocol.h"

#include <stdint.h>

#include "yawns.pb-c.h"


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

int ons_send_pb(struct ons_s *ons, Base* message)
{
    void* buf;

    size_t size = base__get_packed_size(message);

    buf = malloc(size + 1);

    size_t encoded_len = base__pack(message, buf);

    int res = ons_send_msg(ons, buf, encoded_len);

    free(buf);

    return res;
}

RFInfo ons_build_rfinfo(char* band, int channel)
{
    RFInfo info = RFINFO__INIT;

    info.band = band;
    info.channel = channel;

    return info;
}

int ons_send_register(struct ons_s *ons, char* address)
{
    Base base = BASE__INIT;
    Register reg = REGISTER__INIT;

    reg.address = address;

    base.message_case = BASE__MESSAGE_REGISTER;
    base.register_ = &reg;

    return ons_send_pb(ons, &base);
}

int ons_send_field_set(struct ons_s *ons, char* name, char* data_str) {
    Base base = BASE__INIT;
    FieldSet set = FIELD_SET__INIT;

    set.name = name;
    set.data = data_str;

    base.message_case = BASE__MESSAGE_FIELD_SET;
    base.fieldset = &set;

    return ons_send_pb(ons, &base);
}

int ons_send_field_req(struct ons_s *ons, char* name) {
    Base base = BASE__INIT;
    FieldReq req = FIELD_REQ__INIT;

    req.name = name;

    base.message_case = BASE__MESSAGE_FIELD_REQ;
    base.fieldreq = &req;

    return ons_send_pb(ons, &base);
}

int ons_send_deregister(struct ons_s *ons, char* address)
{
    Base base = BASE__INIT;
    Deregister dereg = DEREGISTER__INIT;

    dereg.address = address;

    return ons_send_pb(ons, &base);
}

int ons_send_packet(struct ons_s *ons, char* band, int32_t channel, uint8_t *data, uint16_t length)
{
    Base base = BASE__INIT;
    Packet packet = PACKET__INIT;
    RFInfo info = RFINFO__INIT;

    packet.data.len = length;
    packet.data.data = data;

    info.band = band;
    info.channel = channel;
    packet.info = &info;

    base.message_case = BASE__MESSAGE_PACKET;
    base.packet = &packet;

    return ons_send_pb(ons, &base);
}

int ons_send_rssi_req(struct ons_s *ons, char* band, int channel)
{
    Base base = BASE__INIT;
    RSSIReq req = RSSIREQ__INIT;

    RFInfo info = ons_build_rfinfo(band, channel);
    req.info = &info;

    base.message_case = BASE__MESSAGE_RSSI_REQ;
    base.rssireq = &req;

    return ons_send_pb(ons, &base);
}

int ons_send_state_req(struct ons_s *ons, char* band)
{
    Base base = BASE__INIT;
    StateReq req = STATE_REQ__INIT;

    RFInfo info = RFINFO__INIT;
    info.band = band;
    req.info = &info;

    base.message_case = BASE__MESSAGE_STATE_REQ;
    base.statereq = &req;

    return ons_send_pb(ons, &base);
}

int ons_send_start_receive(struct ons_s *ons, char* band, int channel)
{
    Base base = BASE__INIT;
    StateSet stateset = STATE_SET__INIT;

    RFInfo info = ons_build_rfinfo(band, channel);
    stateset.info = &info;

    stateset.state = RFSTATE__RECEIVE;

    base.message_case = BASE__MESSAGE_STATE_SET;
    base.stateset = &stateset;

    return ons_send_pb(ons, &base);
}

int ons_send_idle(struct ons_s *ons, char* band)
{
    Base base = BASE__INIT;
    StateSet stateset = STATE_SET__INIT;

    RFInfo info = RFINFO__INIT;
    info.band = band;

    stateset.info = &info;
    stateset.state = RFSTATE__IDLE;

    base.message_case = BASE__MESSAGE_STATE_SET;
    base.stateset = &stateset;

    return ons_send_pb(ons, &base);
}

int ons_send_sleep(struct ons_s *ons, char* band)
{
    Base base = BASE__INIT;
    StateSet stateset = STATE_SET__INIT;

    RFInfo info = RFINFO__INIT;
    info.band = band;
    stateset.info = &info;
    stateset.state = RFSTATE__SLEEP;

    base.message_case = BASE__MESSAGE_STATE_SET;
    base.stateset = &stateset;

    return ons_send_pb(ons, &base);
}


int ons_send_event(struct ons_s *ons, char* data)
{
    Base base = BASE__INIT;
    Event event = EVENT__INIT;

    event.data = data;

    return ons_send_pb(ons, &base);
}
