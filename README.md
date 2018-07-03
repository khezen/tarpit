# tarpit

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/khezen/tarpit)

* simple HTTP middleware that purposely delays incoming connections
* repeted calls to a given resource increase the delay
* enable TCP keep alive to keep the client from timing out

## example

```golang

package main

import (
    "net/http"
    "github.com/khezen/tarpit"
)

var tarpitMiddleware = tarpit.New(tarpit.DefaultDelay, tarpit.DefaultResetPeriod)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
      if r.Method != http.MethodGet{
         w.WriteHeader(http.StatusMethodNotAllowed)
         return
    }
    w.Write([]byte("OK"))
}

func handleGetMedicine(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet{
         w.WriteHeader(http.StatusMethodNotAllowed)
         return
    }
    err := tarpitMiddleware.Tar(w, r)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(err.Error()))
        return
    }
    w.Write([]byte("Here are your pills"))
}

func main() {
    http.HandleFunc("/drugs-store/v1/health", handleHealthCheck)
    http.HandleFunc("/drugs-store/v1/medicine", handleGetMedicine)

    writeTimeout := time.Hour
    err := tarpit.ListenAndServe(":80", nil, writeTimeout)
    if err != nil {
        panic(err)
    }
}
```
