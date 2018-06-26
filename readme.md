# tarpit [![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/khezen/tarpit)

* simple HTTP middleware that purposely delays incoming connections
* repeted calls to a given resource increase the delay
* sends one byte of response every few seconds to keep the client from timing out