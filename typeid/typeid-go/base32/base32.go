package base32

// Encoding and Decoding code based on the go implementation of ulid
// found at: https://github.com/oklog/ulid
// (Copyright 2016 The Oklog Authors)
// Modifications made available under the same license as the original

import (
	"errors"
)

const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

func Encode(src [16]byte) string {
	dst := make([]byte, 26)
	// Optimized unrolled loop ahead.

	// 10 byte timestamp
	dst[0] = alphabet[(src[0]&224)>>5]
	dst[1] = alphabet[src[0]&31]
	dst[2] = alphabet[(src[1]&248)>>3]
	dst[3] = alphabet[((src[1]&7)<<2)|((src[2]&192)>>6)]
	dst[4] = alphabet[(src[2]&62)>>1]
	dst[5] = alphabet[((src[2]&1)<<4)|((src[3]&240)>>4)]
	dst[6] = alphabet[((src[3]&15)<<1)|((src[4]&128)>>7)]
	dst[7] = alphabet[(src[4]&124)>>2]
	dst[8] = alphabet[((src[4]&3)<<3)|((src[5]&224)>>5)]
	dst[9] = alphabet[src[5]&31]

	// 16 bytes of entropy
	dst[10] = alphabet[(src[6]&248)>>3]
	dst[11] = alphabet[((src[6]&7)<<2)|((src[7]&192)>>6)]
	dst[12] = alphabet[(src[7]&62)>>1]
	dst[13] = alphabet[((src[7]&1)<<4)|((src[8]&240)>>4)]
	dst[14] = alphabet[((src[8]&15)<<1)|((src[9]&128)>>7)]
	dst[15] = alphabet[(src[9]&124)>>2]
	dst[16] = alphabet[((src[9]&3)<<3)|((src[10]&224)>>5)]
	dst[17] = alphabet[src[10]&31]
	dst[18] = alphabet[(src[11]&248)>>3]
	dst[19] = alphabet[((src[11]&7)<<2)|((src[12]&192)>>6)]
	dst[20] = alphabet[(src[12]&62)>>1]
	dst[21] = alphabet[((src[12]&1)<<4)|((src[13]&240)>>4)]
	dst[22] = alphabet[((src[13]&15)<<1)|((src[14]&128)>>7)]
	dst[23] = alphabet[(src[14]&124)>>2]
	dst[24] = alphabet[((src[14]&3)<<3)|((src[15]&224)>>5)]
	dst[25] = alphabet[src[15]&31]

	return string(dst)
}

// Byte to index table for O(1) lookups when unmarshaling.
// We use 0xFF as sentinel value for invalid indexes.
var dec = [...]byte{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x01,
	0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0A, 0x0B, 0x0C,
	0x0D, 0x0E, 0x0F, 0x10, 0x11, 0xFF, 0x12, 0x13, 0xFF, 0x14,
	0x15, 0xFF, 0x16, 0x17, 0x18, 0x19, 0x1A, 0xFF, 0x1B, 0x1C,
	0x1D, 0x1E, 0x1F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
}

func Decode(s string) ([]byte, error) {
	if len(s) != 26 {
		return nil, errors.New("invalid length")
	}

	v := []byte(s)
	// Check if all the characters are part of the expected base32 character set.
	if dec[v[0]] == 0xFF ||
		dec[v[1]] == 0xFF ||
		dec[v[2]] == 0xFF ||
		dec[v[3]] == 0xFF ||
		dec[v[4]] == 0xFF ||
		dec[v[5]] == 0xFF ||
		dec[v[6]] == 0xFF ||
		dec[v[7]] == 0xFF ||
		dec[v[8]] == 0xFF ||
		dec[v[9]] == 0xFF ||
		dec[v[10]] == 0xFF ||
		dec[v[11]] == 0xFF ||
		dec[v[12]] == 0xFF ||
		dec[v[13]] == 0xFF ||
		dec[v[14]] == 0xFF ||
		dec[v[15]] == 0xFF ||
		dec[v[16]] == 0xFF ||
		dec[v[17]] == 0xFF ||
		dec[v[18]] == 0xFF ||
		dec[v[19]] == 0xFF ||
		dec[v[20]] == 0xFF ||
		dec[v[21]] == 0xFF ||
		dec[v[22]] == 0xFF ||
		dec[v[23]] == 0xFF ||
		dec[v[24]] == 0xFF ||
		dec[v[25]] == 0xFF {
		return nil, errors.New("invalid base32 character")
	}

	id := make([]byte, 16)

	// 6 bytes timestamp (48 bits)
	id[0] = (dec[v[0]] << 5) | dec[v[1]]
	id[1] = (dec[v[2]] << 3) | (dec[v[3]] >> 2)
	id[2] = (dec[v[3]] << 6) | (dec[v[4]] << 1) | (dec[v[5]] >> 4)
	id[3] = (dec[v[5]] << 4) | (dec[v[6]] >> 1)
	id[4] = (dec[v[6]] << 7) | (dec[v[7]] << 2) | (dec[v[8]] >> 3)
	id[5] = (dec[v[8]] << 5) | dec[v[9]]

	// 10 bytes of entropy (80 bits)
	id[6] = (dec[v[10]] << 3) | (dec[v[11]] >> 2) // First 4 bits are the version
	id[7] = (dec[v[11]] << 6) | (dec[v[12]] << 1) | (dec[v[13]] >> 4)
	id[8] = (dec[v[13]] << 4) | (dec[v[14]] >> 1) // First 2 bits are the variant
	id[9] = (dec[v[14]] << 7) | (dec[v[15]] << 2) | (dec[v[16]] >> 3)
	id[10] = (dec[v[16]] << 5) | dec[v[17]]
	id[11] = (dec[v[18]] << 3) | dec[v[19]]>>2
	id[12] = (dec[v[19]] << 6) | (dec[v[20]] << 1) | (dec[v[21]] >> 4)
	id[13] = (dec[v[21]] << 4) | (dec[v[22]] >> 1)
	id[14] = (dec[v[22]] << 7) | (dec[v[23]] << 2) | (dec[v[24]] >> 3)
	id[15] = (dec[v[24]] << 5) | dec[v[25]]

	return id, nil
}
