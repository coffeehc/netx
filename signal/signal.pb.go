// Code generated by protoc-gen-go.
// source: signal.proto
// DO NOT EDIT!

/*
Package signal is a generated protocol buffer package.

It is generated from these files:
	signal.proto

It has these top-level messages:
	Signal
	Header
*/
package signal

import proto "github.com/golang/protobuf/proto"
import "fmt"
import "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Signal struct {
	// 信令值,uint32
	Signal *uint32 `protobuf:"varint,1,req,name=signal" json:"signal,omitempty"`
	// 序列号
	Sequence *int64 `protobuf:"varint,2,opt,name=sequence" json:"sequence,omitempty"`
	// 版本号
	Version *uint32 `protobuf:"varint,3,opt,name=version,def=1" json:"version,omitempty"`
	// 信令头扩展
	Headers []*Header `protobuf:"bytes,4,rep,name=headers" json:"headers,omitempty"`
	// 信令内容
	Data             []byte `protobuf:"bytes,5,opt,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *Signal) Reset()         { *m = Signal{} }
func (m *Signal) String() string { return proto.CompactTextString(m) }
func (*Signal) ProtoMessage()    {}

const Default_Signal_Version uint32 = 1

func (m *Signal) GetSignal() uint32 {
	if m != nil && m.Signal != nil {
		return *m.Signal
	}
	return 0
}

func (m *Signal) GetSequence() int64 {
	if m != nil && m.Sequence != nil {
		return *m.Sequence
	}
	return 0
}

func (m *Signal) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return Default_Signal_Version
}

func (m *Signal) GetHeaders() []*Header {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *Signal) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type Header struct {
	// 头定义
	Key *string `protobuf:"bytes,1,req,name=key" json:"key,omitempty"`
	// 值
	Value            *string `protobuf:"bytes,2,req,name=value" json:"value,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Header) Reset()         { *m = Header{} }
func (m *Header) String() string { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()    {}

func (m *Header) GetKey() string {
	if m != nil && m.Key != nil {
		return *m.Key
	}
	return ""
}

func (m *Header) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

func init() {
	proto.RegisterType((*Signal)(nil), "signal.Signal")
	proto.RegisterType((*Header)(nil), "signal.Header")
}
