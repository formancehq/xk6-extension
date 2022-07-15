package main

import (
	"context"
	"go.k6.io/xk6"
)

func main() {
	builder := xk6.Builder{
		K6Version: "v0.39.0",
		Extensions: []xk6.Dependency{
			{
				PackagePath: "github.com/numary/xk6-extension/k6_openapi3_extension",
				Version:     "main",
			},
		},
	}
	err := builder.Build(context.Background(), "./k6")
	if err != nil {
		panic(err)
	}
}
