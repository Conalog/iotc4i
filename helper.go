package iotc4i

import "encoding/binary"

func ByteListToInteger(fieldBytes []byte) interface{} {
	switch len(fieldBytes) {
	case 1:
		return int(fieldBytes[0])
	case 2:
		return int(binary.LittleEndian.Uint16(fieldBytes))
	case 3, 4:
		// Treat 3 or 4 byte fields as 32-bit integers
		return int(binary.LittleEndian.Uint32(append(fieldBytes, make([]byte, 4-len(fieldBytes))...)))
	default:
		return fieldBytes // Keep as byte slice if more than 4 bytes
	}
}
