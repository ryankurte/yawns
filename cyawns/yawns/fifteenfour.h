#ifndef FIFTEEN_FOUR_H
#define FIFTEEN_FOUR_H

#ifdef __cplusplus
extern "C" {
#endif

// Fifteen Four MAC header
struct fifteen_four_header_s {
    uint8_t frame_ctl1;     //!< Frame control byte one
    uint8_t frame_ctl2;     //!< Frame control byte 2
    uint8_t seq;            //!< Sequence number
    uint16_t pan;           //!< PAN id
    uint16_t dest;          //!< Destination address
    uint16_t src;           //!< Source address
} __attribute((packed));

#define FIFTEEN_FOUR_DEFAULT_HEADER(pan_id, src_addr, dest_addr, sequence) { \
    .frame_ctl1 = 0x61, \
    .frame_ctl2 = 0x88, \
    .seq        = sequence,  \
    .pan        = pan_id,  \
    .dest       = dest_addr, \
    .src        = src_addr   \
}

#ifdef __cplusplus
}
#endif

#endif