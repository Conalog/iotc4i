# IoT C4I

## Description

IoT C4I package provides a simple way to collect data from IoT devices and send it to a server. Still in development. (C4I stands for Command, Control, Communications, Computers, and Intelligence.)

## Features

- [x] Collect data from IoT devices
- [x] Send command to IoT devices
- [ ] Expose route to send data to server
- [ ] Expose route to receive commands from server

## Assumptions

- IoT data receiver receives data from IoT devices via serial port
- Each data is fixed length and uses COBS encoding to separate data

## Sample Code

See `examples/example.go` for details.

## Field data specification

```json
{
    "fields": [
        {
            "name": "RSSI",
            "startIdx": 0,
            "endIdx": 0
        },
        {
            "name": "ProductVersion",
            "startIdx": 7,
            "endIdx": 8
        },
        {
            "startIdx": 48,
            "endIdx": 48,
            "zerofill": true  
        },
        {
            "name": "FieldExample",
            "startIdx": 49,
            "endIdx": 49,
            "zerofill": true
        }
    ]
}
```

- `name`: Field name to be used in the data, omit if not needed
- `startIdx`: Start index of the field in the data
- `endIdx`: End index of the field in the data
- `zerofill`: If true, this field will be zero-filled when hashing

## License

MIT

## Author

Kwangmin Kim (kmkim@conalog.com)