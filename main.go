package k6_openapi3_extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/openapi", new(Module))
}

type Request struct {
	URL     string      `js:"url"`
	Method  string      `js:"method"`
	Body    string      `js:"body"`
	Headers http.Header `js:"headers"`
}

type Response struct {
	StatusCode int         `js:"statusCode"`
	Body       string      `js:"body"`
	Headers    http.Header `js:"headers"`
}

type Validator struct {
	doc    *openapi3.T
	router routers.Router
}

func (v *Validator) Validate(request Request, response Response) error {

	url, err := url.Parse(request.URL)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequest(request.Method, request.URL, bytes.NewBufferString(request.Body))
	if err != nil {
		return fmt.Errorf("invalid request: %s", err)
	}
	httpRequest.Header = request.Headers

	route, pathParams, err := v.router.FindRoute(httpRequest)
	if err != nil {
		return err
	}

	options := &openapi3filter.Options{
		IncludeResponseStatus: true,
		MultiError:            true,
		AuthenticationFunc:    openapi3filter.NoopAuthenticationFunc,
	}
	input := &openapi3filter.RequestValidationInput{
		Request:     httpRequest,
		PathParams:  pathParams,
		QueryParams: url.Query(),
		Route:       route,
		Options:     options,
	}

	err = openapi3filter.ValidateRequest(context.Background(), input)
	if err != nil {
		return err
	}

	err = openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: input,
		Status:                 response.StatusCode,
		Header:                 response.Headers,
		Body:                   ioutil.NopCloser(bytes.NewBufferString(response.Body)),
		Options:                options,
	})
	if err != nil {
		return err
	}

	return nil
}

type Module struct{}

func (*Module) XValidator(spec map[string]interface{}) *Validator {

	data, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, _ := loader.LoadFromData(data)

	err = doc.Validate(ctx)
	if err != nil {
		panic(err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}
	return &Validator{
		doc:    doc,
		router: router,
	}
}
