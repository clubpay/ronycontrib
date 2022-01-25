package swagger

import (
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/ronaksoft/ronykit/desc"
	"github.com/ronaksoft/ronykit/std/bundle/rest"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type swaggerGen struct {
	s       *spec.Swagger
	tagName string
}

func NewSwagger(title, ver, desc string) *swaggerGen {
	sg := &swaggerGen{
		s: &spec.Swagger{},
	}
	sg.s.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Description: desc,
			Title:       title,
			Version:     ver,
		},
	}
	sg.s.Schemes = []string{"http", "https"}
	sg.s.Swagger = "2.0"

	return sg
}

func (sg *swaggerGen) WithTag(tagName string) *swaggerGen {
	sg.tagName = tagName

	return sg
}
func (sg swaggerGen) Generate(w io.Writer, services ...desc.Service) error {
	for _, s := range services {
		addTag(sg.s, s)
		for _, c := range s.Contracts {
			sg.addOperation(sg.s, s.Name, c)
		}
	}

	swaggerJSON, err := sg.s.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = w.Write(swaggerJSON)

	return err
}

func (sg swaggerGen) addOperation(swag *spec.Swagger, serviceName string, c desc.Contract) {
	if swag.Paths == nil {
		swag.Paths = &spec.Paths{
			Paths: map[string]spec.PathItem{},
		}
	}

	inType := reflect.TypeOf(c.Input)
	outType := reflect.TypeOf(c.Output)

	opID := c.Name
	op := spec.NewOperation(opID).
		RespondsWith(
			http.StatusOK,
			spec.NewResponse().
				WithSchema(
					spec.RefProperty(fmt.Sprintf("#/definitions/%s", outType.Elem().Name())),
				),
		).
		WithTags(serviceName).
		WithProduces("application/json").
		WithConsumes("application/json")

	for _, sel := range c.Selectors {
		restSel, ok := sel.(rest.Selector)
		if !ok {
			continue
		}

		sg.setInput(op, restSel.Path, inType)
		addDefinition(swag, inType)
		addDefinition(swag, outType)

		restPath := replacePath(restSel.Path)
		pathItem := swag.Paths.Paths[restPath]
		switch strings.ToUpper(restSel.Method) {
		case rest.MethodGet:
			pathItem.Get = op
		case rest.MethodDelete:
			pathItem.Delete = op
		case rest.MethodPost:
			op.AddParam(
				spec.BodyParam(
					inType.Name(),
					spec.RefProperty(fmt.Sprintf("#/definitions/%s", inType.Name())),
				),
			)
			pathItem.Post = op
		case rest.MethodPut:
			op.AddParam(
				spec.BodyParam(
					inType.Name(),
					spec.RefProperty(fmt.Sprintf("#/definitions/%s", inType.Name())),
				),
			)
			pathItem.Put = op
		case rest.MethodPatch:
			op.AddParam(
				spec.BodyParam(
					inType.Name(),
					spec.RefProperty(fmt.Sprintf("#/definitions/%s", inType.Name())),
				),
			)
			pathItem.Patch = op
		}
		swag.Paths.Paths[restPath] = pathItem
	}
}
func (sg *swaggerGen) setInput(op *spec.Operation, path string, inType reflect.Type) {
	if inType.Kind() == reflect.Ptr {
		inType = inType.Elem()
	}

	var pathParams = make([]string, 0)
	for _, pp := range strings.Split(path, "/") {
		if !strings.HasPrefix(pp, ":") {
			continue
		}
		pathParam := strings.TrimPrefix(pp, ":")
		pathParams = append(pathParams, pathParam)
	}

	for i := 0; i < inType.NumField(); i++ {
		fName := inType.Field(i).Tag.Get(sg.tagName)
		if fName == "" {
			continue
		}
		found := false
		for _, pathParam := range pathParams {
			if fName == pathParam {
				found = true
			}
		}

		if found {
			op.AddParam(
				setParameter(
					spec.PathParam(fName).
						AsRequired().
						NoEmptyValues(),
					inType.Field(i).Type,
				),
			)
		} else {
			op.AddParam(
				setParameter(
					spec.QueryParam(fName).
						AsRequired().
						NoEmptyValues(),
					inType.Field(i).Type,
				),
			)
		}
	}
}
func addTag(swag *spec.Swagger, s desc.Service) {
	swag.Tags = append(
		swag.Tags,
		spec.NewTag(s.Name, "", nil),
	)
}
func addDefinition(swag *spec.Swagger, rType reflect.Type) {
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}

	if swag.Definitions == nil {
		swag.Definitions = map[string]spec.Schema{}
	}

	def := spec.Schema{}
	def.Typed("object", "")

	for i := 0; i < rType.NumField(); i++ {
		f := rType.Field(i)
		fType := f.Type
		fName := f.Name
		var wrapFunc func(schema *spec.Schema) spec.Schema
		switch fType.Kind() {
		case reflect.Ptr:
			fType = fType.Elem()
			wrapFunc = func(schema *spec.Schema) spec.Schema {
				return *schema
			}
		case reflect.Slice:
			wrapFunc = func(item *spec.Schema) spec.Schema {
				return *spec.ArrayProperty(item)
			}
			fType = fType.Elem()
		default:
			wrapFunc = func(schema *spec.Schema) spec.Schema {
				return *schema
			}
		}

		switch fType.Kind() {
		case reflect.String:
			def.SetProperty(fName, wrapFunc(spec.StringProperty()))
		case reflect.Int8, reflect.Uint8:
			def.SetProperty(fName, wrapFunc(spec.ArrayProperty(spec.Int8Property())))
		case reflect.Int32, reflect.Uint32:
			def.SetProperty(fName, wrapFunc(spec.Int32Property()))
		case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64:
			def.SetProperty(fName, wrapFunc(spec.Int64Property()))
		case reflect.Float32:
			def.SetProperty(fName, wrapFunc(spec.Float32Property()))
		case reflect.Float64:
			def.SetProperty(fName, wrapFunc(spec.Float64Property()))
		case reflect.Struct:
			def.SetProperty(fName, wrapFunc(spec.RefProperty(fmt.Sprintf("#/definitions/%s", fType.Name()))))
			addDefinition(swag, fType)
		default:
			fmt.Println(f.Type.Kind())
			def.SetProperty(fName, wrapFunc(spec.StringProperty()))
		}
	}

	swag.Definitions[rType.Name()] = def
}
func setParameter(p *spec.Parameter, t reflect.Type) *spec.Parameter {
	kind := t.Kind()
	switch kind {
	case reflect.Slice:
		switch t.Elem().Kind() {
		case reflect.String:
			p.Typed("string", kind.String())
		case reflect.Float64, reflect.Float32:
			p.Typed("number", kind.String())
		case reflect.Int8, reflect.Uint8:
			p.Typed("integer", "int8")
		case reflect.Int32, reflect.Uint32:
			p.Typed("integer", "int32")
		case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
			p.Typed("integer", "int64")
		default:
			return nil
		}
	case reflect.String:
		p.Typed("string", kind.String())
	case reflect.Float64, reflect.Float32:
		p.Typed("number", kind.String())
	case reflect.Int8, reflect.Uint8:
		p.Typed("integer", "int8")
	case reflect.Int32, reflect.Uint32:
		p.Typed("integer", "int32")
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		p.Typed("integer", "int64")
	default:
		return nil
	}

	return p
}
func replacePath(path string) string {
	sb := strings.Builder{}
	for idx, p := range strings.Split(path, "/") {
		if idx > 0 {
			sb.WriteRune('/')
		}
		if strings.HasPrefix(p, ":") {
			sb.WriteRune('{')
			sb.WriteString(p[1:])
			sb.WriteRune('}')
		} else {
			sb.WriteString(p)
		}
	}

	return sb.String()
}
