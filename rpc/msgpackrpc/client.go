// client
package msgpackrpc

import (
	"fmt"
	"io"
	"math/rand"
	"msgpackgo"
)

func EncodeClientRequest(method string, args interface{}) ([]byte, error) {
	bytes, err := msgpackgo.Marshal(args)
	if err != nil {
		bytes = null
	}
	c := &rpcRequest{
		Method: method,
		Params: bytes,
		Id:     int64(rand.Int63()),
	}
	return msgpackgo.Marshal(c)
}

func DecodeClientResponse(r io.Reader, reply interface{}) error {
	var c rpcResponse
	if err := msgpackgo.NewDecoder(r).Decode(&c); err != nil {
		return err
	}
	if c.Error != "" {
		return fmt.Errorf(c.Error)
	}
	return msgpackgo.Unmarshal(c.Result, reply)
}
