// channle.go
package coffeenet

type Handler interface {
	Active(context *Context)
	Exception(context *Context, err error)
	Read(context *Context, data interface{})
	Close(context *Context)
}
