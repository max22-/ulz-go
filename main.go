package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s input output\n", os.Args[0])
		return
	}
	input, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer input.Close()
	output, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	decoded, err := decode(bufio.NewReader(input))
	if err != nil {
		output.Close()
		os.Remove(os.Args[2])
		fmt.Fprintln(os.Stderr, err)
		return
	}
	n, err := output.Write(decoded)
	if err != nil {
		output.Close()
		os.Remove(os.Args[2])
		fmt.Fprintln(os.Stderr, err)
		return
	} else if n != len(decoded) {
		fmt.Fprintln(os.Stderr, "failed to write output file")
	}
	output.Close()
}

func decode(input *bufio.Reader) ([]byte, error) {
	var output []byte
	i := 0
	for {
		i++
		byte1, err := input.ReadByte()
		if err != nil { // End of input
			return output, nil
		}
		if byte1&0x80 == 0 { // LIT
			length := int(byte1 + 1)
			buffer := make([]byte, length)
			_, err := io.ReadFull(input, buffer)
			if err != nil {
				return nil, err
			}
			output = append(output, buffer...)
		} else if (byte1&0xc0)>>6 == 2 { // CPY1
			length := int((byte1 & 0x3f) + 4)
			offset, err := input.ReadByte()
			if err != nil {
				return nil, err
			}
			offset += 1
			output, err = cpy(output, length, offset)
			if err != nil {
				return nil, err
			}
		} else if (byte1&0xc0)>>6 == 3 { // CPY2
			params := make([]byte, 2)
			_, err := io.ReadFull(input, params)
			if err != nil {
				return nil, err
			}
			length := int(params[0] | (byte1 & 0x3f))
			offset, err := input.ReadByte()
			if err != nil {
				return nil, err
			}
			offset += 1
			output, err = cpy(output, length, offset)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("invalid opcode")
		}
	}
}

func cpy(buffer []byte, length int, offset byte) ([]byte, error) {
	if int(offset) > len(buffer) {
		return nil, errors.New("offset + length - 1 > len(buffer)")
	}
	start := len(buffer) - int(offset)
	end := start + length
	for i := start; i < end; i++ {
		buffer = append(buffer, buffer[i])
	}
	return buffer, nil
}
