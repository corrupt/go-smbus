/*
	Package smbus provides go bindings for the SMBus (System Management Bus) kernel interface
	SMBus is a subset of i2c suitable for a large number of devices
	Original domentation : https://www.kernel.org/doc/Documentation/i2c/smbus-protocol
*/
package smbus

/*
#include "i2c-dev.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	i2c_SLAVE = 0x0703
)

// Base type. Wraps a bus device and an address
type SMBus struct {
	bus  *os.File
	addr byte
}

// Factory method for SMBus
func New(bus uint, address byte) (*SMBus, error) {
	smb := &SMBus{bus: nil}
	err := smb.Bus_open(bus)
	if err != nil {
		return nil, err
	}
	err = smb.Set_addr(address)
	if err != nil {
		return nil, err
	}
	return smb, nil
}

// Opens a new bus file with a given index. Will return an error if a bus is already open
func (smb *SMBus) Bus_open(bus uint) error {

	if smb.bus != nil {
		return errors.New("Can only open one bus at at time")
	}
	path := fmt.Sprintf("/dev/i2c-%d", bus)
	//f, err := os.OpenFile(path, os.O_RDWR, 0600)
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	smb.bus = f
	return nil
}

// Closes an open bus file
func (smb *SMBus) Bus_close() error {
	err := smb.bus.Close()
	if err != nil {
		return err
	} else {
		smb.bus = nil
		return nil
	}
}

// Set the device bus address to a value between 0x00 and 0x77
func (smb *SMBus) Set_addr(addr byte) error {
	if smb.addr != addr {
		if err := ioctl(smb.bus.Fd(), i2c_SLAVE, uintptr(addr)); err != nil {
			return err
		}
		smb.addr = addr
	}
	return nil
}

func ioctl(fd, cmd, arg uintptr) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// Sends a single bit to the device, at the place of the Rd/Wr bit.
func (smb SMBus) Write_quick(value byte) error {
	smb.Set_addr(smb.addr)
	_, err := C.i2c_smbus_write_quick(C.int(smb.bus.Fd()), C.__u8(value))
	return err
}

// Reads a single byte from a device, without specifying a device
// register. Some devices are so simple that this interface is enough;
// for others, it is a shorthand if you want to read the same register
// as in the previous SMBus command.
func (smb SMBus) Read_byte() (byte, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_read_byte(C.int(smb.bus.Fd()))
	if err != nil {
		ret = 0
	}
	return byte(ret & 0x0FF), err
}

// This operation is the reverse of Receive Byte: it sends a single
// byte to a device. See Receive Byte for more information.
func (smb SMBus) Write_byte(value byte) error {
	smb.Set_addr(smb.addr)
	_, err := C.i2c_smbus_write_byte(C.int(smb.bus.Fd()), C.__u8(value))
	return err
}

// Reads a single byte from a device, from a designated register.
// The register is specified through the cmd byte
func (smb SMBus) Read_byte_data(cmd byte) (byte, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_read_byte_data(C.int(smb.bus.Fd()), C.__u8(cmd))
	if err != nil {
		ret = 0
	}
	return byte(ret & 0x0FF), err
}

// Writes a single byte to a device, to a designated register. The
// register is specified through the cmd byte. This is the opposite
// of the Read Byte operation.
func (smb SMBus) Write_byte_data(cmd, value byte) error {
	smb.Set_addr(smb.addr)
	_, err := C.i2c_smbus_write_byte_data(C.int(smb.bus.Fd()), C.__u8(cmd), C.__u8(value))
	return err
}

// This operation is very like Read Byte; again, data is read from a
// device, from a designated register that is specified through the cmd
// byte. But this time, the data is a complete word (16 bits).
func (smb *SMBus) Read_word_data(cmd byte) (uint16, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_read_word_data(C.int(smb.bus.Fd()), C.__u8(cmd))
	if err != nil {
		ret = 0
	}
	return uint16(ret & 0x0FFFF), err
}

// This is the opposite of the Read Word operation. 16 bits
// of data is written to a device, to the designated register that is
// specified through the cmd byte.
func (smb SMBus) Write_word_data(cmd byte, value uint16) error {
	smb.Set_addr(smb.addr)
	_, err := C.i2c_smbus_write_word_data(C.int(smb.bus.Fd()), C.__u8(cmd), C.__u16(value))
	return err
}

// This command selects a device register (through the cmd byte), sends
// 16 bits of data to it, and reads 16 bits of data in return.
func (smb SMBus) Process_call(cmd byte, value uint16) (uint16, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_process_call(C.int(smb.bus.Fd()), C.__u8(cmd), C.__u16(value))
	if err != nil {
		ret = 0
	}
	return uint16(ret & 0x0FFFF), err
}

// This command reads a block of up to 32 bytes from a device, from a
// designated register that is specified through the cmd byte. The amount
// of data in byte is specified by the length of the buf slice.
// To read 4 bytes of data, pass a slice created like this: make([]byte, 4)
func (smb SMBus) Read_block_data(cmd byte, buf []byte) (int, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_read_block_data(
		C.int(smb.bus.Fd()),
		C.__u8(cmd),
		(*C.__u8)(unsafe.Pointer(&buf[0])),
	)
	return int(ret), err
}

// The opposite of the Block Read command, this writes up to 32 bytes to
// a device, to a designated register that is specified through the
// cmd byte. The amount of data is specified by the lengts of buf.
func (smb SMBus) Write_block_data(cmd byte, buf []byte) (int, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_write_block_data(C.int(smb.bus.Fd()), C.__u8(cmd), C.__u8(len(buf)), ((*C.__u8)(&buf[0])))
	return int(ret), err
}

// Block read method for devices without SMBus support. Uses plain i2c interface
func (smb SMBus) Read_i2c_block_data(cmd byte, buf []byte) (int, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_read_i2c_block_data(C.int(smb.bus.Fd()), C.__u8(cmd), C.__u8(len(buf)), ((*C.__u8)(&buf[0])))
	return int(ret), err
}

// Block write method for devices without SMBus support. Uses plain i2c interface
func (smb SMBus) Write_i2c_block_data(cmd byte, buf []byte) (int, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_write_i2c_block_data(C.int(smb.bus.Fd()), C.__u8(cmd), C.__u8(len(buf)), ((*C.__u8)(&buf[0])))
	return int(ret), err
}

// This command selects a device register (through the cmd byte), sends
// 1 to 31 bytes of data to it, and reads 1 to 31 bytes of data in return.
func (smb SMBus) Block_process_call(cmd byte, buf []byte) ([]byte, error) {
	smb.Set_addr(smb.addr)
	ret, err := C.i2c_smbus_block_process_call(C.int(smb.bus.Fd()), C.__u8(cmd), C.__u8(len(buf)), ((*C.__u8)(&buf[0])))
	if err != nil {
		return nil, err
	} else {
		return buf[:ret], nil
	}
}
