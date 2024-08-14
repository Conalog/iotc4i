package iotc4i

import (
	"bytes"
	"errors"
	"fmt"
)

func EncodeCOBS(data []byte, delimiter byte) []byte {
	if len(data) == 0 {
		return []byte{delimiter}
	}

	var result []byte
	index := 0
	result = append(result, 0) // Placeholder for the first length byte

	for i, b := range data {
		if b == delimiter {
			result[index] = byte(i - index + 1)
			index = len(result)
			result = append(result, 0)
		} else {
			result = append(result, b)
		}
	}

	// Set the final length byte
	result[index] = byte(len(data) - index + 1)

	// Add the delimiter at the end
	result = append(result, delimiter)

	return result
}
func DecodeCOBS(encoded []byte, delimiter byte) ([]byte, error) {
	if len(encoded) == 0 {
		return nil, errors.New("encoded data is empty in COBS decoding")
	}

	var decoded bytes.Buffer
	index := 0

	for index < len(encoded) {
		length := int(encoded[index])
		if length == 0 {
			return nil, errors.New("zero byte encountered in COBS decoding")
		}

		// Read next (length-1) bytes into the output
		nextIndex := index + length
		if nextIndex > len(encoded) {
			return nil, errors.New("invalid COBS length")
		}

		decoded.Write(encoded[index+1 : nextIndex])

		// If length is less than 255, append the delimiter byte
		if length < 255 && nextIndex < len(encoded) {
			decoded.WriteByte(delimiter)
		}

		index = nextIndex
	}

	return decoded.Bytes(), nil
}

func (c *C4iHub) DecodeData(data []byte) ([]byte, error) {
	decoded, err := DecodeCOBS(data, c.MessageDelim)
	if err != nil {
		return nil, err
	}
	if len(decoded) != c.MessageSize {
		return nil, fmt.Errorf("message delimeter found but COBS decoded message size is incorrect")
	}
	return decoded, nil
}
