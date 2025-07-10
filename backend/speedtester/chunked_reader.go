package speedtester

import (
	"io"
	"time"
)

type ChunkedZeroReader struct {
	*ZeroReader
	chunkSize    int
	delayBetween time.Duration
	lastRead     time.Time
}

func NewChunkedZeroReader(totalSize int, chunkSize int, delayBetween time.Duration) *ChunkedZeroReader {
	return &ChunkedZeroReader{
		ZeroReader:   NewZeroReader(totalSize),
		chunkSize:    chunkSize,
		delayBetween: delayBetween,
		lastRead:     time.Now(),
	}
}

func (r *ChunkedZeroReader) Read(p []byte) (n int, err error) {
	if r.delayBetween > 0 && time.Since(r.lastRead) < r.delayBetween {
		time.Sleep(r.delayBetween - time.Since(r.lastRead))
	}
	
	maxRead := len(p)
	if r.chunkSize > 0 && maxRead > r.chunkSize {
		maxRead = r.chunkSize
	}
	
	if maxRead < len(p) {
		n, err = r.ZeroReader.Read(p[:maxRead])
	} else {
		n, err = r.ZeroReader.Read(p)
	}
	
	r.lastRead = time.Now()
	return n, err
}

// BufferedReader 提供带缓冲的读取器，适用于VLESS
type BufferedReader struct {
	reader    io.Reader
	buffer    []byte
	bufferPos int
	bufferLen int
}

func NewBufferedReader(reader io.Reader, bufferSize int) *BufferedReader {
	return &BufferedReader{
		reader: reader,
		buffer: make([]byte, bufferSize),
	}
}

func (r *BufferedReader) Read(p []byte) (n int, err error) {
	if r.bufferLen > r.bufferPos {
		n = copy(p, r.buffer[r.bufferPos:r.bufferLen])
		r.bufferPos += n
		return n, nil
	}
	
	r.bufferPos = 0
	r.bufferLen, err = r.reader.Read(r.buffer)
	if err != nil {
		return 0, err
	}
	
	n = copy(p, r.buffer[:r.bufferLen])
	r.bufferPos = n
	return n, nil
}