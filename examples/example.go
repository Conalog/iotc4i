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

	return desiredHash == calculatedHash
}
func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	log.Level(zerolog.InfoLevel)

	hub, err := iotc4i.NewC4iHub("COM62", 460800, 56, 0xCF, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create hub")
		return
	}
	if err := hub.Connect(true, true); err != nil {
		log.Error().Err(err).Msg("Failed to connect")
		return
	}

	dataChan := make(chan []byte, 8192)
	commandChan := make(chan []byte, 50000)
	stopChan := make(chan struct{})
	errorChan := make(chan error)
	warningChan := make(chan error)

	if err := hub.Start(dataChan, commandChan, stopChan, errorChan, warningChan); err != nil {
		fmt.Println("Failed to start processing:", err)
		return
	}

	go func() {
		for {
			select {
			case data := <-dataChan:
				// TODO: Add your version selection here
				byteStartIdx := 7
				byteEndIdx := 8
				productVersion := iotc4i.ByteListToInteger(data[byteStartIdx : byteEndIdx+1]).(int)

				// TODO: Add your field specification loading here
				// Example: Read field specifications from a file
				//				specData, err := iotc4i.ReadFieldSpecificationsFromFile(fmt.Sprintf("specs/%d.json", productVersion))
				// Example: Read field specifications from a server
				specData, err := iotc4i.ReadFieldSpecificationsFromServer(fmt.Sprintf("http://127.0.0.1:9000/specs/%d.json", productVersion), nil)
				if err != nil {
					log.Warn().Err(err).Msg("Error reading field specifications")
					continue
				}
				parsed, err := hub.ParseDataWithSpecification(data, specData)
				if err != nil {
					log.Warn().Err(err).Msg("Error parsing data")
					continue
				}

				// TODO: Add your custom processing here
				if !IsHashValid(parsed) {
					log.Warn().Interface("Parsed", parsed).Msg("Hash Fail")
				} else {
					log.Info().Interface("Parsed", parsed).Msg("Hash OK")
				}
			case err := <-errorChan:
				log.Error().Err(err).Msg("Error processing data")
			case err := <-warningChan:
				log.Warn().Err(err)
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
