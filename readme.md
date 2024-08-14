# IoT C4I

## Description

IoT C4I package provides a simple way to collect data from IoT devices and send it to a server. Still in development. (C4I stands for Command, Control, Communications, Computers, and Intelligence.)

## Features

- [x] Collect data from IoT devices
- [ ] Expose route to send data to server
- [ ] Expose route to receive commands from server

## Assumptions

- IoT data receiver receives data from IoT devices via serial port
- Each data is fixed length and ends with a same delimiter

## Sample Code

See `examples/example.go` for details.

## License

MIT

## Author

Kwangmin Kim (kmkim@conalog.com)