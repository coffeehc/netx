package protocol

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
)

type lengthFieldProtocol struct {
	lengthFieldLength int
	buf               *bytes.Buffer
	length            int64
}

//NewLengthFieldProtocol cteate a LengthField Protocol implement
func NewLengthFieldProtocol(lengthFieldLength int) netx.Protocol {
	if lengthFieldLength != 1 && lengthFieldLength != 2 && lengthFieldLength != 4 && lengthFieldLength != 8 {
		panic("设置的字段长度必须是1,2,4,8,否则协议无法生效")
	}
	p := new(lengthFieldProtocol)
	p.lengthFieldLength = lengthFieldLength
	p.buf = bytes.NewBuffer(nil)
	return p
}

func (lp *lengthFieldProtocol) reset() {
	lp.length = 0
	lp.buf.Reset()
}

func (lp *lengthFieldProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.([]byte); ok {
		length := len(v)
		if length <= 0 {
			logger.Warn("本次发送数据为空,忽略本次数据发送")
			return
		}
		var sendData []byte
		switch lp.lengthFieldLength {
		case 1:
			if length >= 256 {
				logger.Error("发送的数据大于255,丢弃本次数据发送")
				return
			}
			sendData = []byte{byte(length)}
			break
		case 2:
			if length >= 65536 {
				logger.Error("发送的数据大于65536,丢弃本次数据发送")
				return
			}
			sendData = make([]byte, 2)
			binary.BigEndian.PutUint16(sendData, uint16(length))
			break
		case 4:
			sendData = make([]byte, 4)
			binary.BigEndian.PutUint32(sendData, uint32(length))
			break
		case 8:
			sendData = make([]byte, 8)
			binary.BigEndian.PutUint64(sendData, uint64(length))
			break
		default:
			logger.Error("设置了一个错误的字段长度,%d,丢弃本次数据", lp.lengthFieldLength)
			return
		}
		sendData = append(sendData, v...)
		data = sendData
	}
	chain.Process(cxt, connContext, data)
}

func (lp *lengthFieldProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.([]byte); ok {
		if len(v) == 0 {
			logger.Warn("读取的数据为空")
			return
		}
		if lp.length == 0 {
			lengthSize := lp.lengthFieldLength - lp.buf.Len()
			dataLength := len(v)
			if dataLength < lengthSize {
				lp.buf.Write(v)
				return
			}
			if lengthSize > 0 { //不可能有出现0的情况
				lp.buf.Write(v[:lengthSize])
			} else {
				logger.Debug("出现了不可能的情况:lengthSize=%d", lengthSize)
			}
			switch lp.lengthFieldLength {
			case 1:
				lp.length = int64(lp.buf.Bytes()[0])
			case 2:
				lp.length = int64(binary.BigEndian.Uint16(lp.buf.Bytes()))
			case 4:
				lp.length = int64(binary.BigEndian.Uint32(lp.buf.Bytes()))
			case 8:
				lp.length = int64(binary.BigEndian.Uint64(lp.buf.Bytes()))
			}
			lp.buf.Reset()
			v = v[lengthSize:dataLength]
		}
		curLength := int64(lp.buf.Len())
		lastLength := lp.length - curLength
		dataLength := int64(len(v))
		if dataLength < lastLength {
			lp.buf.Write(v)
			return
		}
		lp.buf.Write(v[:lastLength])
		result := make([]byte, lp.length)
		copy(result, lp.buf.Bytes())
		chain.Process(cxt, connContext, result)
		lp.reset()
		if dataLength > lastLength {
			lp.Decode(cxt, connContext, chain, v[lastLength:])
		}
	} else {
		logger.Debug("不能失败")
		chain.Process(cxt, connContext, data)
	}
}

func (lp *lengthFieldProtocol) EncodeDestroy() {}

func (lp *lengthFieldProtocol) DecodeDestroy() {}
