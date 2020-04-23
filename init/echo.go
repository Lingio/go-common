package init

import (
	"context"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/ziflex/lecho/v2"
	"os"
)

func NewEchoServerWithLingioStdConfig(swagger *openapi3.Swagger) *echo.Echo {
	e := echo.New()
	lechologger := lecho.New(os.Stdout, lecho.WithTimestamp(), lecho.WithCaller())
	e.Use(lecho.Middleware(lecho.Config{Logger: lechologger})) // log all requests
	e.Use(echomiddleware.BodyLimit("1M"))                      // limit request body size
	e.Use(echomiddleware.CORS())
	//e.Use(echomiddleware.Gzip())

	// Set up request validation
	options := &middleware.Options{Options: *openapi3filter.DefaultOptions,}
	options.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return nil
	}
	e.Use(oapiRequestValidatorWithOptions(swagger, options)) // check all requests against the OpenAPI schema

	// Init Prometheus
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	return e
}

func oapiRequestValidatorWithOptions(swagger *openapi3.Swagger, options *middleware.Options) echo.MiddlewareFunc {
	router := openapi3filter.NewRouter().WithSwagger(swagger)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/metrics" {
				return next(c)
			}
			err := middleware.ValidateRequestFromContext(c, router, options)
			if err != nil {
				return err
			}
			return next(c)
		}
	}
}
