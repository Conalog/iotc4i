package iotc4i

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
)

type C4iHub struct {
	PortName     string
	BaudRate     int
	MessageSize  int
	MessageDelim byte

	// Internal fields
	serialPort     serial.Port
	dataBufferSize int
	readBufferSize int
	delayAfterRead time.Duration
}

// NewC4iHub creates a new C4iHub instance with the given parameters.
// messageSize is the size of the message payload excluding the COBS encoding stuff.
// messageDelim is the COBS encoded message delimiter.
func NewC4iHub(portName string, baudRate int, messageSize int, messageDelim byte) *C4iHub {
	return &C4iHub{
		PortName:     portName,
		BaudRate:     baudRate,
		MessageSize:  messageSize,
		MessageDelim: messageDelim,

		serialPort:     nil,
		dataBufferSize: 65535,
		readBufferSize: 1024,
		delayAfterRead: 10 * time.Millisecond,
	}
}

// Connect opens the serial port connection.
func (c *C4iHub) Connect() error {
	mode := &serial.Mode{
		BaudRate: c.BaudRate,
	}
	port, err := serial.Open(c.PortName, mode)
	if err != nil {
		return err
	}
	c.serialPort = port
	c.serialPort.ResetInputBuffer()
	return nil
}

// Disconnect closes the serial port connection.
func (c *C4iHub) Disconnect() error {
	if c.serialPort == nil {
		return fmt.Errorf("serial port not connected")
	}
	err := c.serialPort.Close()
	if err != nil {
		return err
	}
	c.serialPort = nil
	return nil
}

// ProcessingLoop reads from the serial port and processes the data.
func (c *C4iHub) ProcessingLoop(dataChan chan<- []byte, stopChan <-chan struct{}, errorChan chan error) error {
	if c.serialPort == nil {
		return fmt.Errorf("serial port not connected")
	}

	dataBuffer := NewCircularQueue[byte](c.dataBufferSize)

	packetSize := c.MessageSize + 1
	parseBuffer := make([]byte, packetSize)
	readCount := 0

	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				readBuffer := make([]byte, c.readBufferSize)
				n, err := c.serialPort.Read(readBuffer)

				// Delay after read
				time.Sleep(c.delayAfterRead)
				if err != nil {
					log.Error().Err(err).Msg("Error reading from serial port")
					c.serialPort.Close()
					c.serialPort = nil
					errorChan <- err
					return
				}

				// append readBuffer to dataBuffer
				for i := 0; i < n; i++ {
					dataBuffer.Enqueue(readBuffer[i])
				}

				// check if dataBuffer might have a remaining packet, at least in size
				if readCount+dataBuffer.Size() >= packetSize && dataBuffer.Size() > 0 {
					for dataBuffer.Size() > 0 && readCount <= packetSize {
						// check if the first byte is the message delimeter
						readByte, err := dataBuffer.Dequeue()

						// This should not happen
						if err != nil {
							log.Error().Err(err).Msg("Error dequeuing from data buffer")
							errorChan <- err
							return
						}

						if readByte == c.MessageDelim {
							if readCount == packetSize {
								// Successfully read a packet
								dataChan <- parseBuffer
								readCount = 0
							} else {
								log.Warn().Int("readCount", readCount).Hex("buffer", parseBuffer[:readCount]).Msg("Message delimeter found but COBS encoded message size is incorrect")
								readCount = 0
							}
						} else {
							if readCount == packetSize {
								log.Warn().Int("readCount", readCount).Hex("buffer", parseBuffer[:readCount]).Msg("Message delimeter not found")
								readCount = 0
							} else {
								// append to parseBuffer
								parseBuffer[readCount] = readByte
								readCount++
							}

						}
					}
					if readCount > packetSize {
						log.Warn().Int("readCount", readCount).Hex("buffer", parseBuffer[:readCount]).Msg("Message buffer full")
						readCount = 0
					}
				}
			}
		}
	}()

	return nil
}

// Starts the data processing loop.
func (c *C4iHub) Start(dataChan chan<- []byte, stopChan <-chan struct{}, errorChan chan error) error {
	if c.serialPort == nil {
		return fmt.Errorf("serial port not connected")
	}

	rawDataChan := make(chan []byte)

	go c.ProcessingLoop(rawDataChan, stopChan, errorChan)

	go func() {
		for rawData := range rawDataChan {
			decodedData, err := c.DecodeData(rawData)
			if err != nil {
				log.Warn().Err(err).Msg("Error processing data")
				continue
			}
			dataChan <- decodedData
		}
	}()
	return nil
}
