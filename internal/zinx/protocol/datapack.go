package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

const headerLen = 8

// DataPack 负责消息的封包与包头解码。
type DataPack struct {
	maxPacketSize uint32
}

// NewDataPack 创建一个带有消息大小限制的封包器。
func NewDataPack(maxPacketSize uint32) *DataPack {
	return &DataPack{maxPacketSize: maxPacketSize}
}

// HeaderLen 返回固定包头长度。
func (p *DataPack) HeaderLen() uint32 {
	return headerLen
}

// Pack 将消息编码为“数据长度 + 消息 ID + 数据”。
func (p *DataPack) Pack(message *Message) ([]byte, error) {
	if message == nil {
		return nil, errors.New("消息不能为 nil")
	}

	dataLen := len(message.data)
	if uint64(dataLen) > math.MaxUint32 {
		return nil, fmt.Errorf("消息长度 %d 超过 uint32 范围", dataLen)
	}
	if p.maxPacketSize > 0 && uint32(dataLen) > p.maxPacketSize {
		return nil, fmt.Errorf("消息长度 %d 超过上限 %d", dataLen, p.maxPacketSize)
	}

	packet := make([]byte, headerLen+dataLen)
	binary.LittleEndian.PutUint32(packet[0:4], uint32(dataLen))
	binary.LittleEndian.PutUint32(packet[4:8], message.id)
	copy(packet[headerLen:], message.data)
	return packet, nil
}

// UnpackHeader 解码固定长度包头，返回消息 ID 和消息体长度。
func (p *DataPack) UnpackHeader(header []byte) (messageID, dataLen uint32, err error) {
	if len(header) < headerLen {
		return 0, 0, io.ErrUnexpectedEOF
	}

	dataLen = binary.LittleEndian.Uint32(header[0:4])
	messageID = binary.LittleEndian.Uint32(header[4:8])
	if p.maxPacketSize > 0 && dataLen > p.maxPacketSize {
		return 0, 0, fmt.Errorf("消息长度 %d 超过上限 %d", dataLen, p.maxPacketSize)
	}

	return messageID, dataLen, nil
}
