# tarpit

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/khezen/tarpit)

* simple HTTP middleware that purposely delays incoming connections
* repeted calls to a given resource increase the delay
* enable TCP keep alive to keep the client from timing out

One typical use case is to protect authentication from brute force.

## example

The following example apply tarpit based on IP address. It is possible to apply tarpit based on any data provided in the request.

```golang

package main

import (
    "net/http"
    "github.com/khezen/tarpit"
)

var tarpitMiddleware = tarpit.New(tarpit.DefaultFreeCallsCount, tarpit.DefaultDelay, tarpit.DefaultResetPeriod)

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
    ipAddr := r.Header.Get(httpHeaderXForwardedFor)
    err := tarpitMiddleware.Tar(ipAddr, w, r)
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
    writeTimeout := 30*time.Second
    err := tarpit.ListenAndServe(":80", nil, writeTimeout)
    if err != nil {
        panic(err)
    }
}
```
