package middleware

import "github.com/clubpay/ronykit"

type serviceWrap struct {
	svc  ronykit.Service
	pre  ronykit.Handler
	post ronykit.Handler
}

func Wrap(svc ronykit.Service, pre, post ronykit.Handler) *serviceWrap {
	return &serviceWrap{
		svc:  svc,
		pre:  pre,
		post: post,
	}
}

func (s serviceWrap) Name() string {
	return s.svc.Name()
}

func (s serviceWrap) Contracts() []ronykit.Contract {
	return s.svc.Contracts()
}

func (s serviceWrap) PreHandlers() []ronykit.Handler {
	var handlers = make([]ronykit.Handler, 0)
	if s.pre != nil {
		handlers = append(handlers, s.pre)
	}

	return append(handlers, s.svc.PreHandlers()...)
}

func (s serviceWrap) PostHandlers() []ronykit.Handler {
	return append(s.svc.PostHandlers(), s.post)
}
