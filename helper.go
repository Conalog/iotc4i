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

func IntegerToByteList(value uint32, numBytes int) []byte {
	bytes := make([]byte, numBytes)
	switch numBytes {
	case 1:
		bytes[0] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(bytes, uint16(value))
	case 3, 4:
		// Treat 3 or 4 byte fields as 32-bit integers
		binary.LittleEndian.PutUint32(bytes, uint32(value))
	}
	return bytes
}

func NewCommandPayload(payload []byte, delimeter byte) []byte {
	// 4 byte aligned length
	payloadToHash := append(payload, make([]byte, 4-len(payload)%4)...)

	hash := CalculateCRC32(payloadToHash)
	hashByteArray := IntegerToByteList(hash, 4)

	// Append the hash to the payload
	payload = append(payload, hashByteArray...)

	// Encode the payload using COBS
	encoded := EncodeCOBS(payload, delimeter)
	return encoded
}
