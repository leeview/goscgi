package goscgi

import (
	"errors"
	"net"
	"strconv"
	"time"
)

type Header map[string]string

var (
	InvalidHeaderErr  = errors.New("Invalid header")
	InvalidContentErr = errors.New("Invalid content size")
	UnexpectedEndErr  = errors.New("Unexpected end of stream")
)

// http://www.python.ca/scgi/protocol.txt
func ReadHeader(conn net.Conn, settings *Settings) (Header, error) {
	var err error
	const buffSize = 8 // first we read only 8 bytes from which we determine the headerSize
	var buff [buffSize]byte
	var alreadyRead, readCnt int
	for alreadyRead < buffSize {
		conn.SetReadDeadline(time.Now().Add(settings.ReadTimeout))
		if readCnt, err = conn.Read(buff[alreadyRead:]); err != nil {
			return nil, err
		}
		alreadyRead += readCnt
	}
	var idx int
	var headerSize int
	var headerSizeStr string
	for idx = 0; idx < readCnt; idx++ {
		if buff[idx] == ':' {
			headerSizeStr = string(buff[:idx])
			idx++ // skip ':'
			break
		}
	}
	if len(headerSizeStr) == 0 {
		return nil, InvalidHeaderErr
	}

	if int64hs, err := strconv.ParseInt(headerSizeStr, 10, 0); err != nil {
		return nil, err
	} else {
		headerSize = int(int64hs)
	}
	if headerSize <= 0 || headerSize > settings.MaxHeaderSize {
		return nil, InvalidHeaderErr
	}

	headerSize++ // add the final ','
	headerBuff := make([]byte, headerSize)
	alreadyRead = readCnt - idx // alreadyRead := size of data read after ':'
	if alreadyRead > 0 {
		// copy alreadyRead data from the initial buffer to the headerBuff
		copy(headerBuff[:alreadyRead], buff[idx:readCnt])
	}
	for alreadyRead < headerSize {
		conn.SetReadDeadline(time.Now().Add(settings.ReadTimeout))
		if readCnt, err = conn.Read(headerBuff[alreadyRead:]); err != nil {
			return nil, err
		}
		alreadyRead += readCnt
	}

	header := Header{}
	var name string
	nameExpected := true
	idx = 0
	for {
		if idx >= headerSize {
			return nil, UnexpectedEndErr
		}
		if nameExpected && headerBuff[idx] == ',' {
			break // end of header reached
		}
		baseIdx := idx
		for idx < headerSize && headerBuff[idx] != 0 {
			idx++
		}
		if idx >= headerSize {
			return nil, UnexpectedEndErr
		}
		if str := string(headerBuff[baseIdx:idx]); nameExpected {
			name = str
			nameExpected = false
		} else {
			header[name] = str
			nameExpected = true
		}
		idx++ // skip 0
	}
	return header, nil
}
