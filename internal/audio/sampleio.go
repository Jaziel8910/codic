package audio

// DecodeWAVFile reads a WAV file and returns the mono-mixed samples plus the
// sample rate. It is the public entry point used by the CLI's sample tools.
func DecodeWAVFile(path string) ([]float64, int, error) {
	return decodeWAV(path)
}

// ResampleAudio resamples a float buffer from srcRate to dstRate using the
// package's internal resampler.
func ResampleAudio(in []float64, srcRate, dstRate int) []float64 {
	return resample(in, srcRate, dstRate)
}
