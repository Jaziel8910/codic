package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// decodeWAV reads a WAV file and returns a mono float64 buffer normalized to
// [-1, 1] along with the file's sample rate.
func decodeWAV(path string) ([]float64, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var riff [4]byte
	if _, err := io.ReadFull(f, riff[:]); err != nil {
		return nil, 0, fmt.Errorf("wav: %w", err)
	}
	if string(riff[:]) != "RIFF" {
		return nil, 0, fmt.Errorf("wav: not a RIFF file: %s", path)
	}

	// Skip "RIFF" magic (already read), the 4-byte size, and "WAVE".
	if _, err := f.Seek(12, io.SeekStart); err != nil {
		return nil, 0, err
	}

	var (
		numChannels uint16
		sampleRate  int
		bits        uint16
		format      uint16
		data        []byte
	)

	// Walk chunks.
	for {
		var hdr [8]byte
		if _, err := io.ReadFull(f, hdr[:]); err != nil {
			break // EOF
		}
		chunkID := string(hdr[0:4])
		chunkSize := binary.LittleEndian.Uint32(hdr[4:8])

		switch chunkID {
		case "fmt ":
			buf := make([]byte, chunkSize)
			if _, err := io.ReadFull(f, buf); err != nil {
				return nil, 0, fmt.Errorf("wav: bad fmt: %w", err)
			}
			format = binary.LittleEndian.Uint16(buf[0:2])
			numChannels = binary.LittleEndian.Uint16(buf[2:4])
			sampleRate = int(binary.LittleEndian.Uint32(buf[4:8]))
			bits = binary.LittleEndian.Uint16(buf[14:16])
		case "data":
			data = make([]byte, chunkSize)
			if _, err := io.ReadFull(f, data); err != nil {
				return nil, 0, fmt.Errorf("wav: bad data: %w", err)
			}
		default:
			if _, err := f.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, 0, fmt.Errorf("wav: seek: %w", err)
			}
		}
	}

	if len(data) == 0 {
		return nil, 0, fmt.Errorf("wav: no data chunk in %s", path)
	}
	if numChannels == 0 {
		numChannels = 1
	}
	if sampleRate == 0 {
		sampleRate = SampleRate
	}

	samples := pcmToFloat(data, numChannels, bits, format)
	return samples, sampleRate, nil
}

// pcmToFloat decodes interleaved PCM bytes into a mono float64 slice.
func pcmToFloat(data []byte, channels uint16, bits uint16, format uint16) []float64 {
	bytesPerSample := int(bits) / 8
	if bytesPerSample == 0 {
		bytesPerSample = 1
	}
	frames := len(data) / bytesPerSample / int(channels)
	out := make([]float64, 0, frames)
	ch := int(channels)
	isFloat := format == 3 // WAV format 3 = IEEE float

	for i := 0; i < frames; i++ {
		var acc float64
		for c := 0; c < ch; c++ {
			off := (i*ch + c) * bytesPerSample
			var v float64
			switch bits {
			case 8:
				v = (float64(data[off]) - 128) / 128.0
			case 16:
				s := int16(binary.LittleEndian.Uint16(data[off : off+2]))
				v = float64(s) / 32768.0
			case 24:
				b0 := uint32(data[off])
				b1 := uint32(data[off+1])
				b2 := uint32(data[off+2])
				u := (b2 << 16) | (b1 << 8) | b0
				if u&0x800000 != 0 {
					u |= 0xFF000000
				}
				v = float64(int32(u)) / 8388604.0
			case 32:
				if isFloat {
					v = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off : off+4])))
				} else {
					s := int32(binary.LittleEndian.Uint32(data[off : off+4]))
					v = float64(s) / 2147483648.0
				}
			}
			acc += v
		}
		out = append(out, acc/float64(ch))
	}
	return out
}
