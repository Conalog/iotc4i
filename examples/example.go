package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"conalog.com/iotc4i"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func IsHashValid(parsed map[string]interface{}) bool {
	// Check if the fields "DesiredHash" and "CalculatedHash" are present in the parsed data
	if _, ok := parsed["DesiredHash"]; !ok {
		log.Error().Msg("DesiredHash not found in parsed data")
		return false
	}

	if _, ok := parsed["CalculatedHash"]; !ok {
		log.Error().Msg("CalculatedHash not found in parsed data")
		return false
	}

	// Compare the desired and calculated hashes
	desiredHash := parsed["DesiredHash"].(uint32)
	calculatedHash := parsed["CalculatedHash"].(uint32)

	if desiredHash != calculatedHash {
		log.Error().Msg("Hashes do not match")
		return false
	}

	return true
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	log.Level(zerolog.InfoLevel)

	hub := iotc4i.NewC4iHub("COM62", 460800, 56, 0xCF)

	if err := hub.Connect(); err != nil {
		log.Error().Err(err).Msg("Failed to connect")
		return
	}

	dataChan := make(chan []byte)
	stopChan := make(chan struct{})
	errorChan := make(chan error)

	if err := hub.Start(dataChan, stopChan, errorChan); err != nil {
		fmt.Println("Failed to start processing:", err)
		return
	}

	go func() {
		for {
			select {
			case data := <-dataChan:
				log.Trace().Hex("framedData", data).Msg("Data processed")
				// Decode the data using the field specifications
				parsed, err := hub.ParseDataWithSpecification("specs", 7, 8, data)
				if err != nil {
					log.Warn().Err(err).Msg("Error parsing data")
					continue
				}

				// Conalog-specific processing
				addrMsb := parsed["DeviceIdHigh"].(int)
				addrLsb := parsed["DeviceIdLow"].(int)
				deviceId := fmt.Sprintf("%04X%08X", addrMsb, addrLsb)
				parsed["DeviceId"] = deviceId
				delete(parsed, "DeviceIdHigh")
				delete(parsed, "DeviceIdLow")

				parsed["ProductVersionWithoutModel"] = parsed["ProductVersion"].(int) & 0x00FF
				parsed["ModelId"] = parsed["ProductVersion"].(int) >> 8
				// Conalog-specific processing

				if !IsHashValid(parsed) {
					log.Warn().Interface("Parsed", parsed).Msg("Hash Fail")
				} else {
					log.Info().Interface("Parsed", parsed).Msg("Hash OK")
				}
			case err := <-errorChan:
				log.Error().Err(err).Msg("Error processing data")
			}
		}
	}()

	// Gracefully handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	close(stopChan)
	fmt.Println("Shutting down")
}
