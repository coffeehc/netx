// rpc project rpc.go
package rpc

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type Codec interface {
	NewRequest(*http.Request) CodecRequest
}

type CodecRequest interface {
	// Reads the request and returns the RPC method name.
	Method() (string, error)
	// Reads the request filling the RPC method args.
	ReadRequest(interface{}) error
	// Writes the response using the RPC method reply.
	WriteResponse(http.ResponseWriter, interface{})
	// Writes an error produced by the server.
	WriteError(w http.ResponseWriter, status int, err error)
}

func NewServer() *Server {
	return &Server{
		codecs:   make(map[string]Codec),
		services: new(serviceMap),
	}
}

type Server struct {
	codecs   map[string]Codec
	services *serviceMap
}

func (s *Server) RegisterCodec(codec Codec, contentType string) {
	s.codecs[strings.ToLower(contentType)] = codec
}

func (s *Server) RegisterService(receiver interface{}, name string) error {
	return s.services.register(receiver, name)
}

func (s *Server) HasMethod(method string) bool {
	if _, _, err := s.services.get(method); err == nil {
		return true
	}
	return false
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		WriteError(w, 405, "rpc: POST method required, received "+r.Method)
		return
	}
	contentType := r.Header.Get("Content-Type")
	idx := strings.Index(contentType, ";")
	if idx != -1 {
		contentType = contentType[:idx]
	}
	codec := s.codecs[strings.ToLower(contentType)]
	if codec == nil {
		WriteError(w, 415, "rpc: unrecognized Content-Type: "+contentType)
		return
	}
	codecReq := codec.NewRequest(r)
	method, errMethod := codecReq.Method()
	if errMethod != nil {
		codecReq.WriteError(w, 400, errMethod)
		return
	}
	serviceSpec, methodSpec, errGet := s.services.get(method)
	if errGet != nil {
		codecReq.WriteError(w, 400, errGet)
		return
	}
	args := reflect.New(methodSpec.argsType)
	if errRead := codecReq.ReadRequest(args.Interface()); errRead != nil {
		codecReq.WriteError(w, 400, errRead)
		return
	}
	reply := reflect.New(methodSpec.replyType)
	errValue := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(r),
		args,
		reply,
	})
	var errResult error
	errInter := errValue[0].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}
	w.Header().Set("x-content-type-options", "nosniff")
	if errResult == nil {
		codecReq.WriteResponse(w, reply.Interface())
	} else {
		codecReq.WriteError(w, 400, errResult)
	}
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, msg)
}
