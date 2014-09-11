// msgpake project msgpake.go
package msgpackrpc

import (
	"coffeenet/rpc"
	"errors"
	"msgpackgo"
	"net/http"
)

var null = make([]byte, 0)

type rpcRequest struct {
	Method string
	Params []byte
	Id     int64
}

type rpcResponse struct {
	Result []byte
	Error  string
	Id     int64
}

func NewCodec() *Codec {
	return &Codec{}
}

type Codec struct {
}

func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
	req := new(rpcRequest)
	err := msgpackgo.NewDecoder(r.Body).Decode(req)
	r.Body.Close()
	return &CodecRequest{request: req, err: err}
}

type CodecRequest struct {
	request *rpcRequest
	err     error
}

func (c *CodecRequest) Method() (string, error) {
	if c.err == nil {
		return c.request.Method, nil
	}
	return "", c.err
}

func (c *CodecRequest) ReadRequest(args interface{}) error {
	if c.err == nil {
		if c.request.Params != nil {
			c.err = msgpackgo.Unmarshal(c.request.Params, &args)
		} else {
			c.err = errors.New("rpc: method request ill-formed: missing params field")
		}
	}
	return c.err
}

func (c *CodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}) {
	if c.request.Id != 0 {
		bs, err := msgpackgo.Marshal(reply)
		if err != nil {
			bs = null
		}
		res := &rpcResponse{
			Result: bs,
			Id:     c.request.Id,
			Error:  "",
		}
		c.writeServerResponse(w, 200, res)
	}
}

func (c *CodecRequest) WriteError(w http.ResponseWriter, _ int, err error) {
	res := &rpcResponse{
		Result: null,
		Id:     c.request.Id,
		Error:  "",
	}
	if err != nil {
		res.Error = err.Error()
	}
	c.writeServerResponse(w, 400, res)
}

func (c *CodecRequest) writeServerResponse(w http.ResponseWriter, status int, res *rpcResponse) {
	b, err := msgpackgo.Marshal(res)
	//TODO 此处考虑对称加密
	if err == nil {
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/object; charset=utf-8")
		w.Write(b)
	} else {
		rpc.WriteError(w, 400, err.Error())
	}
}
