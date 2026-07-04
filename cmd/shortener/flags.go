package main

import (
    "flag"
)

var flagRunAddr string
var flagRedirectAddr string

func parseFlags() {
    flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
    flag.StringVar(&flagRedirectAddr, "b", "http://localhost:8080", "address and port for redirect url")
    flag.Parse()
}