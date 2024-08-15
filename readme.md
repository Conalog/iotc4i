# IoT C4I

## Description

IoT C4I(Command, Control, Communications and Intelligence + Computer) package provides a simple way to collect data from IoT devices and send it to a server. Make your own IoT data receiver and command sender with this package.

Currently supports `Command` and `Communications`, stil working on rest of the features.

## Requirements

- IoT data receiver receives data from IoT devices via serial port
- Each data is fixed length and uses COBS encoding to separate data

## Features

- [x] Collect data from IoT devices
- [x] Send command to IoT devices
- [ ] Exposes telemetry data
- [ ] Provides device intelligence (e.g. reception rate, etc.)

## Procedure

1. Set port configuration, desired data size
2. Open serial port, start receiving data
3. Command data may be sent to the device (optional)
4. Process received data on the fly
5. Dynamic field parsing is supported (optional, see Field data specification and Sample Code)
6. CRC32 checksum while zero-fill specific fields (optional, see Field data specification and Sample Code)

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

## Sample Code

See `examples/example.go` for details.

## License

MIT

## Author

Kwangmin Kim (<kmkim@conalog.com>)
