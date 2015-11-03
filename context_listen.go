// context_listen
package coffeenet

type ContextListen interface {
	OnActive(context *Context)
	OnClose(context *Context)
	OnException(context *Context, err error)
}

type SimpleContextListen struct {
}

func (this *SimpleContextListen) OnActive(context *Context) {
	//do Nothing
}

func (this *SimpleContextListen) OnClose(context *Context) {
	//do Nothing
}
func (this *SimpleContextListen) OnException(context *Context, err error) {
	//do Nothing
}
