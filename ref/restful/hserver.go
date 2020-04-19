package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})
	http.HandleFunc("/abc", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world ,abc"))
	})
	if err := http.ListenAndServe(":12345", nil); err != nil {
		fmt.Println("start http server fail:", err)
	}
}
