package speedtester

import (
	"io"
	"time"
)

// ChunkedZeroReader 提供分块零字节读取，用于优化VLESS上传
type ChunkedZeroReader struct {
	*ZeroReader
	chunkSize    int
	delayBetween time.Duration
	lastRead     time.Time
}

// NewChunkedZeroReader 创建一个分块零字节读取器
func NewChunkedZeroReader(totalSize int, chunkSize int, delayBetween time.Duration) *ChunkedZeroReader {
	return &ChunkedZeroReader{
		ZeroReader:   NewZeroReader(totalSize),
		chunkSize:    chunkSize,
		delayBetween: delayBetween,
		lastRead:     time.Now(),
	}
}

func (r *ChunkedZeroReader) Read(p []byte) (n int, err error) {
	// 如果需要延迟，等待一段时间
	if r.delayBetween > 0 && time.Since(r.lastRead) < r.delayBetween {
		time.Sleep(r.delayBetween - time.Since(r.lastRead))
	}
	
	// 限制每次读取的大小
	maxRead := len(p)
	if r.chunkSize > 0 && maxRead > r.chunkSize {
		maxRead = r.chunkSize
	}
	
	// 使用父类的Read方法，但限制大小
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

// NewBufferedReader 创建一个带缓冲的读取器
func NewBufferedReader(reader io.Reader, bufferSize int) *BufferedReader {
	return &BufferedReader{
		reader: reader,
		buffer: make([]byte, bufferSize),
	}
}

func (r *BufferedReader) Read(p []byte) (n int, err error) {
	// 如果缓冲区有数据，先从缓冲区读取
	if r.bufferLen > r.bufferPos {
		n = copy(p, r.buffer[r.bufferPos:r.bufferLen])
		r.bufferPos += n
		return n, nil
	}
	
	// 缓冲区为空，填充缓冲区
	r.bufferPos = 0
	r.bufferLen, err = r.reader.Read(r.buffer)
	if err != nil {
		return 0, err
	}
	
	// 从新填充的缓冲区读取
	n = copy(p, r.buffer[:r.bufferLen])
	r.bufferPos = n
	return n, nil
}