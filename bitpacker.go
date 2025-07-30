package gobitpacker

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Unpack reads bits from a byte slice into the fields of the struct pointed to by v.
// The struct fields must be exported and can be of type uint8, uint16, uint32, or uint64.
// Fields should be tagged with a `bits:"N"` tag where N is the number of bits to read.
func Unpack(data []byte, v any) error {
	// Get the value that v points to
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return errors.New("destination must be a non-nil pointer to a struct")
	}

	// Dereference the pointer to get the struct value
	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to struct, got %s", elem.Kind())
	}

	// Keep track of our position in the bit stream
	bitOffset := 0

	// Iterate over all fields of the struct
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)
		fieldValue := elem.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get the bits tag
		bitsTag, hasBits := field.Tag.Lookup("bits")
		if !hasBits {
			continue // Skip fields without bits tag
		}

		// Parse the number of bits to read
		bitCount, err := strconv.Atoi(bitsTag)
		if err != nil || bitCount <= 0 {
			return fmt.Errorf("invalid bits value '%s' for field %s", bitsTag, field.Name)
		}

		// Calculate how many bytes we need to read
		bitsAvailable := (len(data)*8 - bitOffset)
		if bitsAvailable < bitCount {
			return fmt.Errorf("not enough bits to read field %s (needed %d, available %d)",
				field.Name, bitCount, bitsAvailable)
		}

		// Read the bits
		value, err := readBits(data, bitOffset, bitCount)
		if err != nil {
			return fmt.Errorf("error reading bits for field %s: %w", field.Name, err)
		}

		// Set the field value based on its type
		switch fieldValue.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldValue.SetUint(uint64(value))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldValue.SetInt(int64(value))
		default:
			return fmt.Errorf("unsupported field type %s for field %s",
				fieldValue.Kind(), field.Name)
		}

		// Move the bit offset
		bitOffset += bitCount
	}

	return nil
}

// readBits reads a number of bits from a byte slice starting at a given bit offset
func readBits(data []byte, offset, count int) (uint64, error) {
	if count > 64 {
		return 0, errors.New("cannot read more than 64 bits at once")
	}

	var result uint64
	bitsRemaining := count

	for bitsRemaining > 0 {
		// Calculate current byte and bit positions
		bytePos := offset / 8
		bitPos := offset % 8

		if bytePos >= len(data) {
			return 0, errors.New("not enough data to read bits")
		}

		// Calculate how many bits we can read from the current byte
		bitsInByte := 8 - bitPos
		bitsToRead := bitsInByte
		if bitsToRead > bitsRemaining {
			bitsToRead = bitsRemaining
		}

		// Extract the bits from the current byte
		mask := byte((1 << bitsToRead) - 1)
		bits := (data[bytePos] >> (8 - bitPos - bitsToRead)) & mask

		// Add the bits to our result
		result = (result << bitsToRead) | uint64(bits)

		// Update counters
		offset += bitsToRead
		bitsRemaining -= bitsToRead
	}

	return result, nil
}
