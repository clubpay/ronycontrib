package middleware

import "github.com/clubpay/ronykit"

type serviceWrap struct {
	svc  ronykit.Service
	pre  ronykit.HandlerFunc
	post ronykit.HandlerFunc
}

func Wrap(svc ronykit.Service, pre, post ronykit.HandlerFunc) *serviceWrap {
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

func (s serviceWrap) PreHandlers() []ronykit.HandlerFunc {
	var handlers = make([]ronykit.HandlerFunc, 0)
	if s.pre != nil {
		handlers = append(handlers, s.pre)
	}

	return append(handlers, s.svc.PreHandlers()...)
}

func (s serviceWrap) PostHandlers() []ronykit.HandlerFunc {
	return append(s.svc.PostHandlers(), s.post)
}
