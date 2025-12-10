package initflow

type Step func(*Context) error
