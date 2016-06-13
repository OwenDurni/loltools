package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func FetchUrl(loc string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "Fetch: %s\n", loc)
	res, err := http.Get(loc)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
