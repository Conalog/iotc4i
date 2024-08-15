package iotc4i

import (
	"fmt"
	"time"

	"go.bug.st/serial"
)

type C4iHub struct {
	PortName         string
	BaudRate         int
	MessageSize      int
	MessageDelimiter byte

	// Optional parameters
	DataBufferSize int
	ReadBufferSize int
	DelayAfterRead time.Duration
	ReadTimeout    time.Duration

	// Internal fields
	serialPort serial.Port
}
type C4iHubOptions struct {
	DataBufferSize *int
	ReadBufferSize *int
	DelayAfterRead *time.Duration
	ReadTimeout    *time.Duration
}

// NewC4iHub creates a new C4iHub instance with the given parameters.
// messageSize is the size of the message payload excluding the COBS encoding stuff.
// messageDelimiter is the COBS encoded message delimiter.
// options
// - ReadBufferSize: ReadBufferSize > 0 (default : 1024)
// - DataBufferSize: DataBufferSize > 0 && DataBufferSize >= ReadBufferSize (default : 65535)
// - DelayAfterRead: DelayAfterRead > 0 (default : 10ms)
// - ReadTimeout: ReadTimeout > 0 (default : 100ms)
func NewC4iHub(portName string, baudRate int, messageSize int, messageDelimiter byte, options *C4iHubOptions) (*C4iHub, error) {
	hub := &C4iHub{
		PortName:         portName,
		BaudRate:         baudRate,
		MessageSize:      messageSize,
		MessageDelimiter: messageDelimiter,

		serialPort:     nil,
		ReadBufferSize: 1024,
		DataBufferSize: 65535,
		DelayAfterRead: 10 * time.Millisecond,
		ReadTimeout:    100 * time.Millisecond,
	}
	if options != nil {
		if options.ReadBufferSize != nil {
			if *options.ReadBufferSize <= 0 {
				return nil, fmt.Errorf("read buffer size must be greater than 0")
			}
			hub.ReadBufferSize = *options.ReadBufferSize
		}
		if options.DataBufferSize != nil {
			if *options.DataBufferSize <= 0 {
				return nil, fmt.Errorf("data buffer size must be greater than 0")
			}
			if *options.DataBufferSize < hub.ReadBufferSize {
				return nil, fmt.Errorf("data buffer size must be greater than read buffer size")
			}
			hub.DataBufferSize = *options.DataBufferSize
		}
		if options.DelayAfterRead != nil {
			if *options.DelayAfterRead <= 0 {
				return nil, fmt.Errorf("delay after read must be greater than 0")
			}
			hub.DelayAfterRead = *options.DelayAfterRead
		}
		if options.ReadTimeout != nil {
			if *options.ReadTimeout <= 0 {
				return nil, fmt.Errorf("read timeout must be greater than 0")
			}
			hub.ReadTimeout = *options.ReadTimeout
		}
	}
	return hub, nil
}

// Connect opens the serial port connection.
func (c *C4iHub) Connect(flushInputBufferAfterConnect bool, flushOutputBufferAfterConnect bool) error {
	mode := &serial.Mode{
		BaudRate: c.BaudRate,
	}
	port, err := serial.Open(c.PortName, mode)
	if err != nil {
		return err
	}
	c.serialPort = port

	if flushInputBufferAfterConnect {
		if err := c.serialPort.ResetInputBuffer(); err != nil {
			return err
		}
	}
	if flushOutputBufferAfterConnect {
		if err := c.serialPort.ResetOutputBuffer(); err != nil {
			return err
		}
	}
	if err := c.serialPort.SetReadTimeout(c.ReadTimeout); err != nil {
		return err
	}
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
func (c *C4iHub) ProcessingLoop(dataChan chan<- []byte, commandChan chan []byte, stopChan <-chan struct{}, errorChan chan error, warningChan chan error) error {
	if c.serialPort == nil {
		return fmt.Errorf("serial port not connected")
	}

	dataBuffer := NewCircularQueue[byte](c.DataBufferSize)

	packetSize := c.MessageSize + 1
	parseBuffer := make([]byte, packetSize)
	readCount := 0

	go func() {
		for {
			select {
			case <-stopChan:
				return
			case commandData := <-commandChan:
				n, err := c.serialPort.Write(commandData)
				if err != nil {
					if err := c.Disconnect(); err != nil {
						errorChan <- err
					}
					return
				}
				if n != len(commandData) {
					warningChan <- fmt.Errorf("command data not fully written")
				}
			default:
				readBuffer := make([]byte, c.ReadBufferSize)
				n, err := c.serialPort.Read(readBuffer)
				if err != nil {
					if err := c.Disconnect(); err != nil {
						errorChan <- err
					}
					return
				}

				// append readBuffer to dataBuffer
				for i := 0; i < n; i++ {
					if err := dataBuffer.Enqueue(readBuffer[i]); err != nil {
						errorChan <- err
						return
					}
				}

				// check if dataBuffer might have a remaining packet, at least in size
				if readCount+dataBuffer.Size() >= packetSize && dataBuffer.Size() > 0 {
					for dataBuffer.Size() > 0 && readCount <= packetSize {
						// check if the first byte is the message delimeter
						readByte, err := dataBuffer.Dequeue()

						// This should not happen
						if err != nil {
							errorChan <- err
							return
						}

						if readByte == c.MessageDelimiter {
							if readCount == packetSize {
								// Successfully read a packet
								dataChan <- parseBuffer
								readCount = 0
							} else {
								warningChan <- fmt.Errorf("message delimeter found but COBS encoded message size is incorrect")
								readCount = 0
							}
						} else {
							if readCount == packetSize {
								warningChan <- fmt.Errorf("message delimeter not found")
								readCount = 0
							} else {
								// append to parseBuffer
								parseBuffer[readCount] = readByte
								readCount++
							}
						}
					}

					// This should not happen
					if readCount > packetSize {
						warningChan <- fmt.Errorf("message buffer full")
						readCount = 0
					}
				}

				// Delay after read
				time.Sleep(c.DelayAfterRead)
			}
		}
	}()

	return nil
}

// Starts the data processing loop.
func (c *C4iHub) Start(dataChan chan<- []byte, commandChan chan []byte, stopChan <-chan struct{}, errorChan chan error, warningChan chan error) error {
	if c.serialPort == nil {
		return fmt.Errorf("serial port not connected")
	}

	rawDataChan := make(chan []byte)

	go c.ProcessingLoop(rawDataChan, commandChan, stopChan, errorChan, warningChan)

	go func() {
		for rawData := range rawDataChan {
			decodedData, err := c.DecodeData(rawData)
			if err != nil {
				warningChan <- err
				continue
			}
			dataChan <- decodedData
		}
	}()
	return nil
}
