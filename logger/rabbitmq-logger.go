package main

import (
	"fmt"
	"net/http"
)

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handle")
	fmt.Println(r.Method)
	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "Sorry, only POST methods are supported.")
		// Future use
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		fmt.Fprintf(w, "Post call ! r.PostFrom = %v\n", r.PostForm)

		hello := r.FormValue("hello")
		sink := r.FormValue("sink")

		fmt.Fprintf(w, "Hello = %s\n", hello)
		fmt.Fprintf(w, "Sink = %s\n", sink)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}
func main() {
	fmt.Println("Server starting !")
	http.HandleFunc("/", handle)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
