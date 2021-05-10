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
		Addr:         ":9977",
		Handler:      &Proxy{Ctx: ctx},
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
