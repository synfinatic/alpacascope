package main

import (
	"github.com/go-resty/resty/v2"
)

type Alpaca struct {
	client *resty.Client
	ascom  *Ascom
}

func New(ascom *Ascom) *Alpaca {
	a := Alpaca{
		client: resty.New(),
		ascom:  ascom,
	}
	return &a
}
