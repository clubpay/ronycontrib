package cors

import (
	"net/http"
	"strings"

	"github.com/ronaksoft/ronycontrib/middleware"
	"github.com/ronaksoft/ronykit"
	"github.com/ronaksoft/ronykit/std/bundle/rest"
	"github.com/valyala/fasthttp"
)

type Config struct {
	AllowedHeaders []string
	AllowedMethods []string
	AllowedOrigins []string
}

type cors struct {
	headers string
	methods string
	origins string
}

func Handler(cfg Config) ronykit.Handler {
	c := &cors{}
	if len(cfg.AllowedOrigins) == 0 {
		c.origins = "*"
	} else {
		c.origins = strings.Join(cfg.AllowedOrigins, ", ")
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = []string{
			"Origin", "Accept", "Content-Type",
			"X-Requested-With", "X-Auth-Tokens", "Authorization",
		}
	}
	c.headers = strings.Join(cfg.AllowedHeaders, ",")
	if len(cfg.AllowedMethods) == 0 {
		c.methods = strings.Join([]string{
			fasthttp.MethodGet, fasthttp.MethodHead, fasthttp.MethodPost,
			fasthttp.MethodPatch, fasthttp.MethodConnect, fasthttp.MethodDelete,
			fasthttp.MethodTrace, fasthttp.MethodOptions,
		}, ", ")
	} else {
		c.methods = strings.Join(cfg.AllowedMethods, ", ")
	}

	return func(ctx *ronykit.Context) {
		rc, ok := ctx.Conn().(ronykit.REST)
		if !ok {
			return
		}

		// ByPass cors (Cross Origin Resource Sharing) check
		if c.origins == "*" {
			rc.Set(HeaderAccessControlAllowOrigin, rc.Get(HeaderOrigin))
		} else {
			rc.Set(HeaderAccessControlAllowOrigin, c.origins)
		}

		rc.Set(HeaderAccessControlRequestMethod, c.methods)

		if rc.GetMethod() == rest.MethodOptions {
			reqHeaders := rc.Get(HeaderAccessControlRequestHeaders)
			if len(reqHeaders) > 0 {
				rc.Set(HeaderAccessControlAllowHeaders, reqHeaders)
			} else {
				rc.Set(HeaderAccessControlAllowHeaders, c.headers)
			}

			ctx.SetStatusCode(http.StatusNoContent)
			ctx.StopExecution()

			return
		}
	}
}

func CORS(cfg Config) ronykit.ServiceWrapper {
	return func(service ronykit.Service) ronykit.Service {
		return middleware.Wrap(service, Handler(cfg), nil)
	}
}
