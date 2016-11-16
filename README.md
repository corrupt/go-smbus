# go-smbus
Go bindings for the System Management Bus (SMBus) kernel interface
This package provides simple bindings for the SMBus interfaces provided by the i2c-dev driver. I wrote this for the Raspberry Pi platform.

**This code is largely untested. I'll happily accept a pull request for any bugs you might find**

## Installation

    go get github.com/corrupt/go-smbus

## Usage

Create an instance of `SMBus` using the factory method. It takes two parameters, the interface index and the bus address. The former is the enumerated device index. If your I2C device is `/dev/i2c-1`, your index is 1.
The latter is the bus address to connect to from 0x00 to 0x77. It can later be changed using the `Set_addr` method.

```go
smb, err := smbus.New(1, 0x68)
if err != nil {
    fmt.Println(err)              
    os.Exit(1)  
}
```

You can now use the SMBus API to write to and read from the bus. All methods evaluate `errno` and return a go error accordingly. Block read/write methods also return the number of read/written bytes.

```go
cmd := 0xD0
val := 0x10
err := smb.Write_byte_data(cmd, val)

buf := make ([]byte, 4)
i, err := smb.Read_i2c_block_data(0xD1, buf)
if err != nil {
    fmt.Println(err)              
    //error handling
}
if i != len(buf) {
    //error handling
}
```

### Changing the Bus Device

The `SMBus` type provides two convenience methods to close the existing device and optionally open a new one

```go
smb.Bus_close()
smb.Bus_open(0x71)
```
