package iotc4i

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"os"
)

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
	defer client.CloseIdleConnections()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	for key, value := range headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data.Fields, nil
}
