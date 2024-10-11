//go:build wireinject
// +build wireinject

package main

import (
	"net/http"

	"github.com/google/wire"
)

func InitializeApp() (*http.Server, func(), error) {
	wire.Build(WireSet)
	return &http.Server{}, func() {}, nil
}
