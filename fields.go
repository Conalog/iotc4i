package iotc4i

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
)

// LoadFieldSpecifications loads field specifications from a JSON file.
func LoadFieldSpecifications(filename string) ([]Field, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data struct {
		Fields []Field `json:"fields"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data.Fields, nil
}

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
			switch len(fieldBytes) {
			case 1:
				result[field.Name] = int(fieldBytes[0])
			case 2:
				result[field.Name] = int(binary.LittleEndian.Uint16(fieldBytes))
			case 3, 4:
				// Treat 3 or 4 byte fields as 32-bit integers
				result[field.Name] = int(binary.LittleEndian.Uint32(append(fieldBytes, make([]byte, 4-len(fieldBytes))...)))
			default:
				result[field.Name] = fieldBytes // Keep as byte slice if more than 4 bytes
			}
		}
	}

	return result, nil
}

func (c *C4iHub) ParseDataWithSpecification(specDataRoot string, specDataByteStart int, specDataByteEnd int, payload []byte) (map[string]interface{}, error) {
	productVersion := 0

	// Extract the product version field from the payload
	fieldBytes := payload[specDataByteStart : specDataByteEnd+1]
	switch len(fieldBytes) {
	case 1:
		productVersion = int(fieldBytes[0])
	case 2:
		productVersion = int(binary.LittleEndian.Uint16(fieldBytes))
	case 3, 4:
		// Treat 3 or 4 byte fields as 32-bit integers
		productVersion = int(binary.LittleEndian.Uint32(append(fieldBytes, make([]byte, 4-len(fieldBytes))...)))
	default:
		return nil, fmt.Errorf("invalid product version field length: %d", len(fieldBytes))
	}

	// Load the field specifications for the product version
	testFields, err := LoadFieldSpecifications(fmt.Sprintf("%s/%d.json", specDataRoot, productVersion))
	if err != nil {
		return nil, err
	}

	// Decode the message payload using the field specifications
	parsed, err := TryDecodeData(payload, testFields)
	if err != nil {
		return nil, err
	}

	hash := binary.LittleEndian.Uint32(payload[c.MessageSize-4 : c.MessageSize])
	calculatedHash := CalculateHashWithZerofill(payload, testFields)

	parsed["DesiredHash"] = hash
	parsed["CalculatedHash"] = calculatedHash
	return parsed, nil
}
