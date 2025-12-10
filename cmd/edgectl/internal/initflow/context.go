package initflow

type Context struct {
	OS        string
	Arch      string
	HasSystemd bool

	COEndpoint string
	SiteID     string
}

func NewContext() *Context {
	return &Context{}
}
