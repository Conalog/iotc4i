package iotc4i

import "hash/crc32"

// CalculateCRC32 calculates the CRC32 checksum of the given data using a specified polynomial.
func CalculateCRC32(data []byte) uint32 {
	// Create a new CRC32 table based on the polynomial
	table := crc32.MakeTable(crc32.IEEE)

	// Calculate the CRC32 checksum
	checksum := crc32.Checksum(data, table)

	return checksum
}

func CalculateHashWithZerofill(data []byte, fields []Field) uint32 {
	zerofilledData := make([]byte, len(data))
	// Zero-fill the specified fields to calculate the CRC32 checksum
	copy(zerofilledData, data)
	for _, field := range fields {
		if field.Zerofill {
			for i := field.StartIdx; i <= field.EndIdx; i++ {
				zerofilledData[i] = 0
			}
		}
	}

	zerofilledData[len(zerofilledData)-4] = 0
	// Calculate the CRC32 checksum of the zero-filled data
	hash := CalculateCRC32(zerofilledData[1 : len(zerofilledData)-4+1])
	return hash
}
