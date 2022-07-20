package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		panic("test program only accept one param")
	}

	reqAddr := os.Args[1]
	if !strings.HasPrefix(reqAddr, "http://") {
		reqAddr = "http://" + reqAddr
	}

	response, err := http.Get(reqAddr)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
