package binary

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"io"
	"strconv"

	"github.com/beito123/binary"
)

const (
	// TriadSize is byte size of Triad
	TriadSize = 3
)

const (
	MinTriad = 0
	MaxTriad = 16777216
)

// ToTriad convert int to Triad.
// but minus is no supported.
// example: 16777217 int -> 0 Triad
func ToTriad(a int) Triad {
	return Triad(a % (MaxTriad + 1))
}

// Triad is 3bytes data for Raknet
// It's used as index counter in Raknet
type Triad uint32

func (t Triad) Bump() Triad {
	return (t % MaxTriad) + 1
}

func (t Triad) Add(d int) (result Triad) {
	result = t + Triad(d)

	if !t.Vaild() {
		panic("constant" + strconv.Itoa(int(result)) + "overflows Triad")
	}

	return result
}

func (t Triad) Sub(d int) Triad {
	return t.Add(-d)
}

func (t Triad) Multi(d int) (result Triad) {
	result = t * ToTriad(d)
	if !t.Vaild() {
		panic("constant" + strconv.Itoa(int(result)) + "overflows Triad")
	}

	return result
}

func (t Triad) Divide(d int) (result Triad) {
	result = t / ToTriad(d)
	if !t.Vaild() {
		panic("constant" + strconv.Itoa(int(result)) + "overflows Triad")
	}

	return result
}

func (t Triad) Remainder(d int) (result Triad) {
	result = t % ToTriad(d)
	if !t.Vaild() {
		return 0
	}

	return result
}

func (t Triad) Vaild() bool {
	return t >= MinTriad && t <= MaxTriad
}

// ReadTriad read Triad value
func ReadTriad(v []byte) Triad {
	return Triad(v[0])<<16 | Triad(v[1])<<8 | Triad(v[2])
}

// WriteTriad write Triad value
func WriteTriad(v Triad) []byte {
	return []byte{
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
}

// ReadLTriad read Triad value as LittleEndian
func ReadLTriad(v []byte) Triad {
	return Triad(v[0]) | Triad(v[1])<<8 | Triad(v[2])<<16
}

// WriteTriad write Triad value as LittleEndian
func WriteLTriad(v Triad) []byte {
	return []byte{
		byte(v),
		byte(v >> 8),
		byte(v >> 16),
	}
}

// ReadETriad read Triad value with error
func ReadETriad(v []byte) (Triad, error) {
	if len(v) < TriadSize {
		return 0, nil
	}

	return ReadTriad(v), nil
}

// ReadELTriad read Triad value as LittleEndian with error
func ReadELTriad(v []byte) (Triad, error) {
	if len(v) < TriadSize {
		return 0, nil
	}

	return ReadLTriad(v), nil
}

// Read reads data into b by order
func Read(reader io.Reader, order RaknetOrder, data interface{}) error {
	size := dataSize(data)

	bytes := make([]byte, size)

	n, err := reader.Read(bytes)
	if err != nil {
		return err
	}

	if n < size {
		return binary.ErrNotEnought
	}

	switch value := data.(type) {
	case *int8:
		*value = order.SByte(bytes)
	case *uint8:
		*value = order.Byte(bytes)
	case *int16:
		*value = order.Short(bytes)
	case *uint16:
		*value = order.UShort(bytes)
	case *Triad:
		*value = order.Triad(bytes)
	case *int32:
		*value = order.Int(bytes)
	case *uint32:
		*value = order.UInt(bytes)
	case *int64:
		*value = order.Long(bytes)
	case *uint64:
		*value = order.ULong(bytes)
	case *float32:
		*value = order.Float(bytes)
	case *float64:
		*value = order.Double(bytes)
	}

	return nil
}

// Write writes the contents of data into buffer by order
func Write(writer io.Writer, order RaknetOrder, data interface{}) error {
	var value []byte
	switch v := data.(type) {
	case int8:
		value = order.PutSByte(v)
	case *int8:
		value = order.PutSByte(*v)
	case uint8:
		value = order.PutByte(v)
	case *uint8:
		value = order.PutByte(*v)
	case int16:
		value = order.PutShort(v)
	case *int16:
		value = order.PutShort(*v)
	case uint16:
		value = order.PutUShort(v)
	case *uint16:
		value = order.PutUShort(*v)
	case Triad:
		value = order.PutTriad(v)
	case *Triad:
		value = order.PutTriad(*v)
	case int32:
		value = order.PutInt(v)
	case *int32:
		value = order.PutInt(*v)
	case uint32:
		value = order.PutUInt(v)
	case *uint32:
		value = order.PutUInt(*v)
	case int64:
		value = order.PutLong(v)
	case *int64:
		value = order.PutLong(*v)
	case uint64:
		value = order.PutULong(v)
	case *uint64:
		value = order.PutULong(*v)
	}

	_, err := writer.Write(value)

	return err
}

// dataSize returns byte size of type
func dataSize(data interface{}) int {
	var size int
	switch data.(type) {
	case int8, *int8, uint8, *uint8:
		size = binary.ByteSize
	case int16, *int16, uint16, *uint16:
		size = binary.ShortSize
	case Triad, *Triad:
		size = TriadSize
	case int32, *int32, uint32, *uint32:
		size = binary.IntSize
	case int64, *int64, uint64, *uint64:
		size = binary.LongSize
	case float32, *float32:
		size = binary.FloatSize
	case float64, *float64:
		size = binary.DoubleSize
	}

	return size
}

// RaknetOrder for Raknet Protocol
type RaknetOrder interface {
	binary.Order

	Triad(v []byte) Triad
	PutTriad(v Triad) []byte
}

// BigEndian .
var BigEndian bigEndian

// LittleEndian .
var LittleEndian littleEndian

type bigEndian struct {
}

func (bigEndian) Byte(v []byte) byte {
	return binary.ReadByte(v)
}

func (bigEndian) PutByte(v byte) []byte {
	return binary.WriteByte(v)
}

func (bigEndian) SByte(v []byte) int8 {
	return binary.ReadSByte(v)
}

func (bigEndian) PutSByte(v int8) []byte {
	return binary.WriteSByte(v)
}

func (bigEndian) Short(v []byte) int16 {
	return binary.ReadShort(v)
}

func (bigEndian) PutShort(v int16) []byte {
	return binary.WriteShort(v)
}

func (bigEndian) UShort(v []byte) uint16 {
	return binary.ReadUShort(v)
}

func (bigEndian) PutUShort(v uint16) []byte {
	return binary.WriteUShort(v)
}

func (bigEndian) Triad(v []byte) Triad {
	return ReadTriad(v)
}

func (bigEndian) PutTriad(v Triad) []byte {
	return WriteTriad(v)
}

func (bigEndian) Int(v []byte) int32 {
	return binary.ReadInt(v)
}

func (bigEndian) PutInt(v int32) []byte {
	return binary.WriteInt(v)
}

func (bigEndian) UInt(v []byte) uint32 {
	return binary.ReadUInt(v)
}

func (bigEndian) PutUInt(v uint32) []byte {
	return binary.WriteUInt(v)
}

func (bigEndian) Long(v []byte) int64 {
	return binary.ReadLong(v)
}

func (bigEndian) PutLong(v int64) []byte {
	return binary.WriteLong(v)
}

func (bigEndian) ULong(v []byte) uint64 {
	return binary.ReadULong(v)
}

func (bigEndian) PutULong(v uint64) []byte {
	return binary.WriteULong(v)
}

func (bigEndian) Float(v []byte) float32 {
	return binary.ReadFloat(v)
}

func (bigEndian) PutFloat(v float32) []byte {
	return binary.WriteFloat(v)
}

func (bigEndian) Double(v []byte) float64 {
	return binary.ReadDouble(v)
}

func (bigEndian) PutDouble(v float64) []byte {
	return binary.WriteDouble(v)
}

type littleEndian struct {
}

func (littleEndian) Byte(v []byte) byte {
	return binary.ReadByte(v)
}

func (littleEndian) PutByte(v byte) []byte {
	return binary.WriteByte(v)
}

func (littleEndian) SByte(v []byte) int8 {
	return binary.ReadSByte(v)
}

func (littleEndian) PutSByte(v int8) []byte {
	return binary.WriteSByte(v)
}

func (littleEndian) Short(v []byte) int16 {
	return binary.ReadLShort(v)
}

func (littleEndian) PutShort(v int16) []byte {
	return binary.WriteLShort(v)
}

func (littleEndian) UShort(v []byte) uint16 {
	return binary.ReadLUShort(v)
}

func (littleEndian) PutUShort(v uint16) []byte {
	return binary.WriteLUShort(v)
}

func (littleEndian) Triad(v []byte) Triad {
	return ReadLTriad(v)
}

func (littleEndian) PutTriad(v Triad) []byte {
	return WriteLTriad(v)
}

func (littleEndian) Int(v []byte) int32 {
	return binary.ReadLInt(v)
}

func (littleEndian) PutInt(v int32) []byte {
	return binary.WriteLInt(v)
}

func (littleEndian) UInt(v []byte) uint32 {
	return binary.ReadLUInt(v)
}

func (littleEndian) PutUInt(v uint32) []byte {
	return binary.WriteLUInt(v)
}

func (littleEndian) Long(v []byte) int64 {
	return binary.ReadLLong(v)
}

func (littleEndian) PutLong(v int64) []byte {
	return binary.WriteLLong(v)
}

func (littleEndian) ULong(v []byte) uint64 {
	return binary.ReadLULong(v)
}

func (littleEndian) PutULong(v uint64) []byte {
	return binary.WriteLULong(v)
}

func (littleEndian) Float(v []byte) float32 {
	return binary.ReadLFloat(v)
}

func (littleEndian) PutFloat(v float32) []byte {
	return binary.WriteLFloat(v)
}

func (littleEndian) Double(v []byte) float64 {
	return binary.ReadLDouble(v)
}

func (littleEndian) PutDouble(v float64) []byte {
	return binary.WriteLDouble(v)
}
