// frame_procotol
package coffeenet

import (
	"bytes"
	"encoding/binary"

	"github.com/coffeehc/logger"
)

type LengthFieldProtocol struct {
	lengthFieldLength int
	buf               *bytes.Buffer
	length            int64
}

func NewLengthFieldProtocol(lengthFieldLength int) *LengthFieldProtocol {
	if lengthFieldLength != 1 && lengthFieldLength != 2 && lengthFieldLength != 4 && lengthFieldLength != 8 {
		panic("设置的字段长度必须是1,2,4,8,否则协议无法生效")
	}
	p := new(LengthFieldProtocol)
	p.lengthFieldLength = lengthFieldLength
	p.buf = bytes.NewBuffer(nil)
	return p
}

func (this *LengthFieldProtocol) reset() {
	this.length = 0
	this.buf.Reset()
}

func (this *LengthFieldProtocol) Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		length := len(v)
		if length <= 0 {
			logger.Error("本次发送数据为空,忽略本次数据发送")
			return
		}
		var sendData []byte
		switch this.lengthFieldLength {
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
			logger.Error("设置了一个错误的字段长度,%d,丢弃本次数据", this.lengthFieldLength)
			return
		}
		sendData = append(sendData, v...)
		warp.FireNextWrite(context, sendData)
	} else {
		warp.FireNextWrite(context, data)
	}
}

func (this *LengthFieldProtocol) Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		if len(v) == 0 {
			logger.Warn("读取的数据为空")
			return
		}
		if this.length == 0 {
			lengthSize := this.lengthFieldLength - this.buf.Len()
			dataLength := len(v)
			if dataLength < lengthSize {
				this.buf.Write(v)
				return
			}
			if lengthSize > 0 { //不可能有出现0的情况
				this.buf.Write(v[:lengthSize])
			} else {
				logger.Debug("出现了不可能的情况:lengthSize=%d", lengthSize)
			}
			switch this.lengthFieldLength {
			case 1:
				this.length = int64(this.buf.Bytes()[0])
			case 2:
				this.length = int64(binary.BigEndian.Uint16(this.buf.Bytes()))
			case 4:
				this.length = int64(binary.BigEndian.Uint32(this.buf.Bytes()))
			case 8:
				this.length = int64(binary.BigEndian.Uint64(this.buf.Bytes()))
			}
			this.buf.Reset()
			v = v[lengthSize:dataLength]
		}
		curLength := int64(this.buf.Len())
		lastLength := this.length - curLength
		dataLength := int64(len(v))
		if dataLength < lastLength {
			this.buf.Write(v)
			return
		}
		this.buf.Write(v[:lastLength])
		result := make([]byte, this.length)
		copy(result, this.buf.Bytes())
		warp.FireNextRead(context, result)
		this.reset()
		if dataLength > lastLength {
			this.Decode(context, warp, v[lastLength:])
		}
	} else {
		logger.Debug("不能失败")
		warp.FireNextRead(context, data)
	}
}
