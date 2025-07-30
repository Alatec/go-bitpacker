package gobitpacker_test

import (
	"reflect"
	"testing"

	gobitpacker "github.com/Alatec/go-bitpacker"
)

func TestUnpack(t *testing.T) {
	type SingleByte struct {
		Field1 uint8 `bits:"8"`
	}
	type MultipleBytes struct {
		Field1 uint8 `bits:"4"`
		Field2 uint8 `bits:"4"`
		Field3 uint8 `bits:"6"`
		Field4 uint8 `bits:"2"`
	}
	type MixedInts struct {
		U8  uint8  `bits:"8"`
		I16 int16  `bits:"8"`
		U32 uint32 `bits:"8"`
	}
	tests := []struct {
		name    string
		data    []byte
		setup   func() any
		want    any
		wantErr bool
	}{
		{
			name: "single byte field",
			data: []byte{0b10101010},
			setup: func() any {
				return &SingleByte{}
			},
			want: &SingleByte{Field1: 0b10101010},
		},
		{
			name: "multiple fields across bytes",
			data: []byte{0b11011010, 0b10101100},
			setup: func() any {
				return &MultipleBytes{}
			},
			want: &MultipleBytes{
				Field1: 0b1101,
				Field2: 0b1010,
				Field3: 0b101011,
				Field4: 0b00,
			},
		},
		{
			name: "mixed integer types",
			data: []byte{0b11011010, 0b10101100, 0b11111111},
			setup: func() any {
				return &MixedInts{}
			},
			want: &MixedInts{
				U8:  0b11011010,
				I16: 0b10101100,
				U32: 0b11111111,
			},
		},
		{
			name: "not enough data",
			data: []byte{0b11011010},
			setup: func() any {
				return &MultipleBytes{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new instance for each test
			dst := tt.setup()

			// Call the function we're testing
			err := gobitpacker.Unpack(tt.data, dst)

			// Check for expected errors
			if (err != nil) != tt.wantErr {
				t.Logf("Unpack() dst = %+v", dst)
				t.Fatalf("Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If we expected an error, we're done
			if tt.wantErr {
				return
			}

			// Compare the result with what we expected
			if !reflect.DeepEqual(dst, tt.want) {
				t.Errorf("Unpack() = %+v, want %+v", dst, tt.want)
			}
		})
	}
}

func TestUnpack_InvalidInput(t *testing.T) {
	t.Run("nil destination", func(t *testing.T) {
		err := gobitpacker.Unpack([]byte{0}, nil)
		if err == nil {
			t.Error("expected error for nil destination")
		}
	})

	t.Run("non-pointer destination", func(t *testing.T) {
		var s struct {
			Field uint8 `bits:"8"`
		}
		err := gobitpacker.Unpack([]byte{0}, s)
		if err == nil {
			t.Error("expected error for non-pointer destination")
		}
	})

	t.Run("non-struct destination", func(t *testing.T) {
		var i int
		err := gobitpacker.Unpack([]byte{0}, &i)
		if err == nil {
			t.Error("expected error for non-struct destination")
		}
	})

	t.Run("invalid bits tag", func(t *testing.T) {
		type testStruct struct {
			Field uint8 `bits:"invalid"`
		}
		err := gobitpacker.Unpack([]byte{0}, &testStruct{})
		if err == nil {
			t.Error("expected error for invalid bits tag")
		}
	})
}
