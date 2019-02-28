package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Data struct {
	Owner         string `json:"owner"`
	Body          string `json:"body"`
	Critical      bool   `json:"critical"`
	CorrelationID string `json:"correlationId"`
}

func handle(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	data.Body = "OK"
	data.Owner = "logger1"

	b, err := json.Marshal(&data)
	if err != nil {
		panic(err)
	}

	fmt.Println("Sending Data...")
	fmt.Print(string(b))
	fmt.Fprint(w, string(b))
}

func main() {
	fmt.Println("Server starting !")
	http.HandleFunc("/", handle)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
