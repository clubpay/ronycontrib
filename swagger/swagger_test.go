package swagger_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/clubpay/ronycontrib/swagger"
	"github.com/clubpay/ronykit/desc"
	"github.com/clubpay/ronykit/std/bundle/fasthttp"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
)

type sampleReq struct {
	X string   `json:"x"`
	Y string   `json:"y"`
	Z int64    `json:"z"`
	W []string `json:"w"`
}

func (x sampleReq) Marshal() ([]byte, error) {
	return json.Marshal(x)
}

type subRes struct {
	Some    string `json:"some"`
	Another []byte `json:"another"`
}

type sampleRes struct {
	Out1        int         `json:"out1"`
	Out2        string      `json:"out2"`
	Sub         subRes      `json:"sub"`
	Subs        []subRes    `json:"subs"`
	EnumericSub EnumericSub `json:"enumericSub"`
}

type EnumericSub string

const (
	EnumericSubX EnumericSub = "X"
	EnumericSubY EnumericSub = "Y"
	EnumericSubZ EnumericSub = "Z"
)

func (x sampleRes) Marshal() ([]byte, error) {
	return json.Marshal(x)
}

type sampleError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (x sampleError) Marshal() ([]byte, error) {
	return json.Marshal(x)
}

var testService = (&desc.Service{
	Name:         "testService",
	PreHandlers:  nil,
	PostHandlers: nil,
}).AddContract(
	desc.NewContract().
		AddSelector(fasthttp.Selector{
			Method: fasthttp.MethodGet,
			Path:   "/some/:x/:y",
		}).
		SetInput(&sampleReq{}).
		SetOutput(&sampleRes{}).
		AddPossibleError(404, "ITEM1", &sampleError{}).
		AddPossibleError(504, "SERVER", &sampleError{}).
		SetHandler(nil),
)

func TestNewSwagger(t *testing.T) {
	Convey("GenerateSwaggerSpec", t, func(c C) {
		sg := swagger.NewSwagger("TestTitle", "v0.0.1", "")
		sg.WithTag("json")

		sb := new(strings.Builder)
		errWrite := sg.WriteTo(sb, testService)
		So(errWrite, ShouldBeEmpty)

		//spec, errMarshal := json.MarshalIndent(json.RawMessage(sb.String()), "", "   ")
		expectedSpec := `{
	"schemes": ["http", "https"],
	"swagger": "2.0",
	"info": {
		"title": "TestTitle",
		"version": "v0.0.1"
	},
	"tags": [{
		"name": "testService"
	}],
	"paths": {
		"/some/{x}/{y}": {
			"get": {
				"consumes": ["application/json"],
				"parameters": [{
					"format": "string",
					"in": "path",
					"name": "x",
					"required": true,
					"type": "string"
				}, {
					"format": "string",
					"in": "path",
					"name": "y",
					"required": true,
					"type": "string"
				}, {
					"format": "int64",
					"in": "query",
					"name": "z",
					"required": true,
					"type": "integer"
				}, {
					"format": "slice",
					"in": "query",
					"name": "w",
					"required": true,
					"type": "string"
				}],
				"produces": ["application/json"],
				"responses": {
					"200": {
						"description": "",
						"schema": {
							"$ref": "#/definitions/sampleRes"
						}
					},
					"404": {
						"description": "Items: ITEM1",
						"schema": {
							"$ref": "#/definitions/sampleError"
						}
					},
					"504": {
						"description": "Items: SERVER",
						"schema": {
							"$ref": "#/definitions/sampleError"
						}
					}
				},
				"tags": ["testService"]
			}
		}
	},
	"definitions": {
		"sampleError": {
			"properties": {
				"code": {
					"format": "int64",
					"type": "integer"
				},
				"description": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"sampleReq": {
			"properties": {
				"w": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"x": {
					"type": "string"
				},
				"y": {
					"type": "string"
				},
				"z": {
					"format": "int64",
					"type": "integer"
				}
			},
			"type": "object"
		},
		"sampleRes": {
			"properties": {
				"enumericSub": {
					"type": "string"
				},
				"out1": {
					"format": "int64",
					"type": "integer"
				},
				"out2": {
					"type": "string"
				},
				"sub": {
					"$ref": "#/definitions/subRes"
				},
				"subs": {
					"items": {
						"$ref": "#/definitions/subRes"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"subRes": {
			"properties": {
				"another": {
					"items": {
						"items": {
							"format": "int8",
							"type": "integer"
						},
						"type": "array"
					},
					"type": "array"
				},
				"some": {
					"type": "string"
				}
			},
			"type": "object"
		}
	}
}`
		So(sb.String(), should.EqualJSON, expectedSpec)
	})
}
