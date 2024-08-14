package iotc4i

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func ReadFieldSpecificationsFromFile(filename string) ([]Field, error) {
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

func ReadFieldSpecificationsFromServer(url string, headers map[string]string) ([]Field, error) {
	var data struct {
		Fields []Field `json:"fields"`
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
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
