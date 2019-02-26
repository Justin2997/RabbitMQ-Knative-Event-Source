package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Data struct {
	Producer string `json:"producer"`
	Body     string `json:"body"`
	Critical bool   `json:"critical"`
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

	fmt.Printf("Thank you %s, call number %s", data.Producer, data.Body)
	fmt.Fprintf(w, "Thanks you %s, call number %s", data.Producer, data.Body)
}

func main() {
	fmt.Println("Server starting !")
	http.HandleFunc("/", handle)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
