package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Hello BuGuai !!! ")
	ctx := context.TODO()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", DefaultProxyConfig.ServerPort),
		Handler:      &Proxy{Ctx: ctx},
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
