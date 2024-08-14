package iotc4i

import (
	"encoding/binary"
	"fmt"
)

// TryDecodeData decodes the given byte slice into a map according to the field specifications.
func TryDecodeData(data []byte, fields []Field) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, field := range fields {
		// Skip processing if the field has no name
		if field.Name == "" {
			continue
		}

		if field.StartIdx > field.EndIdx || field.EndIdx >= len(data) {
			return nil, fmt.Errorf("invalid index range for field: %s", field.Name)
		}

		// Extract the bytes for the current field
		fieldBytes := data[field.StartIdx : field.EndIdx+1]

		// Interpret the fieldBytes as a single integer
		if field.Name != "" {
			result[field.Name] = ByteListToInteger(fieldBytes)
		}
	}

	return result, nil
}

func (c *C4iHub) ParseDataWithSpecification(payload []byte, specData []Field) (map[string]interface{}, error) {
	// Decode the message payload using the field specifications
	parsed, err := TryDecodeData(payload, specData)
	if err != nil {
		return nil, err
	}

	hash := binary.LittleEndian.Uint32(payload[c.MessageSize-4 : c.MessageSize])
	calculatedHash := CalculateHashWithZerofill(payload, specData)

	parsed["DesiredHash"] = hash
	parsed["CalculatedHash"] = calculatedHash
	return parsed, nil
}
