package usbtemp

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

var mode115200 = &serial.Mode {
	BaudRate: 115200,
	DataBits: 8,
}

var mode9600 = &serial.Mode {
	BaudRate: 9600,
	DataBits: 8,
}

type USBtemp struct {
	Name string
	Id string
	SerialNumber string
	port serial.Port
}

func (u *USBtemp) Open(portName string) error {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return fmt.Errorf("enumerating error: %w", err)
	}

	if len(ports) == 0 {
		return fmt.Errorf("no valid ports found")
	}

	var curPort *enumerator.PortDetails
	for _, port := range ports {
		if port.Name == portName {
			curPort = port
			u.Name = port.Name
		}
	}
	if !curPort.IsUSB {
		return fmt.Errorf("%s is not USB", portName)
	}

	u.Id = fmt.Sprintf("%s:%s", curPort.VID, curPort.PID)
	u.SerialNumber = curPort.SerialNumber

	port, err := serial.Open(portName, mode9600)
	if err != nil {
		return fmt.Errorf("failed to open port: %w", err)
	}
	u.port = port
	return nil
}

func (u *USBtemp) Close() error {
	err := u.port.Close()
	return err
}

func (u *USBtemp) Temperature(fahrenheit bool) (float32, error) {
	if err := u.reset(); err != nil {
		return 0, err
	}
	if err := u.write(0xcc); err != nil {
		return 0, err
	}
	if err := u.write(0x44); err != nil {
		return 0, err
	}
	time.Sleep(time.Second)
	if err := u.reset(); err != nil {
		return 0, err
	}
	if err := u.write(0xcc); err != nil {
		return 0, err
	}
	if err := u.write(0xbe); err != nil {
		return 0, err
	}
	tempBytes, err := u.readBytes(9)
	if err != nil {
		return 0, err
	}
	if u.crc8([8]byte(tempBytes[0:8])) != tempBytes[8] {
		return 0, fmt.Errorf("Temperature() failed: CRC validation")
	}
	for i := 2; i < 9; i++ {
		tempBytes[i] = 0
	}
	celsius := float32(binary.LittleEndian.Uint32(tempBytes)) / 16.0
	if fahrenheit {
		return (9 * celsius) / 5 + 32, nil
	} 
	return celsius, nil
}

func (u *USBtemp) Rom() (string, error) {
	if err := u.reset(); err != nil {
		return "", err
	}
	if err := u.write(0x33); err != nil {
		return "", err
	}
	data, err := u.readBytes(8)
	if err != nil {
		return "", fmt.Errorf("Rom() read failed: %w", err)
	}
	return hex.EncodeToString(data), nil
}

func (u *USBtemp) reset() error {
	err := u.port.SetMode(mode9600)
	if err != nil {
		return fmt.Errorf("reset() setmode failed: %w", err)
	}
	wbuff := make([]byte, 1)
	rbuff := make([]byte, 1)

	u.port.Drain()
	wbuff[0] = 0xf0
	u.port.Write(wbuff)

	u.port.Read(rbuff)
	err = u.port.SetMode(mode115200)
	if err != nil {
		return fmt.Errorf("reset() setmode failed: %w", err)
	}

	if len(rbuff) != 1 {
		return fmt.Errorf("reset() failed: invalid reply length")
	}
	if rbuff[0] == 0xf0 {
		return fmt.Errorf("reset() failed: no device present")
	}
	if rbuff[0] == 0x00 {
		return fmt.Errorf("reset() failed: short circuit")
	}

	if 0x10 <= rbuff[0] && rbuff[0] <= 0xe0 {
		return nil
	}

	return fmt.Errorf("reset() failed, presence error: %v", rbuff[0])
}

func (u *USBtemp) writeByte(b byte) (byte, error) {
	var w []byte
	var rbuff []byte
	temprbuff := make([]byte, 8)

	for i := 0; i < 8; i++ {
		if (b & 0x01) != 0 {
			w = append(w, 0xff)
		} else {
			w = append(w, 0x00)
		}
		b = b >> 1
	}

	wnum, err := u.port.Write(w)
	if err != nil {
		return 0x00, fmt.Errorf("writeByte(): %w", err)
	}
	if wnum != len(w) {
		return 0x00, fmt.Errorf("writeByte(): # of bytes written incorrect")
	}
	readBytes := 0;
	for readBytes < 8 {
		read, err := u.port.Read(temprbuff)
		if err != nil {
			return 0x00, fmt.Errorf("writeByte(): %w", err)
		}
		readBytes += read 
		rbuff = append(rbuff, temprbuff[0:read]...)
	}
	
	value := 0
	for _, b := range rbuff {
		value = value >> 1
		if b == 0xff {
			value = value | 0x80
		}
	}
	return byte(value), nil

}

func (u *USBtemp) write (b byte) error {
	bb, err := u.writeByte(b)
	if err != nil {
		return fmt.Errorf("write(): %w", err)
	}
	if bb != b {
		return fmt.Errorf("write(): read byte does not match write")
	}
	return nil
}

func (u *USBtemp) readBytes(numBytes int) ([]byte, error) {
	var x []byte
	for i := 0; i<numBytes; i++ {
		curByte, err := u.writeByte(0xff)
		if err != nil {
			return nil, fmt.Errorf("readBytes(): %w", err)
		}
		x = append(x, curByte)
	}
	return x, nil
}

func (u *USBtemp) crc8(needCheck [8]byte) byte {
	var crc byte = 0x00

	for _, b := range needCheck[0:8] {
		for i:=0; i < 8; i++ {
			mix := (crc ^ b) & 0x01
			crc = crc >> 1
			if mix == 1 {
				crc = crc ^ 0x8c
			}
			b = b >> 1
		}
	}
	return crc
}
