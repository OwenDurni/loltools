package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Returns the contents of the page at `loc` and the HTTP status code of the
// response.
func FetchUrl(loc string) ([]byte, int, error) {
	fmt.Fprintf(os.Stderr, "Fetch: %s\n", loc)
	res, err := http.Get(loc)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}
	return data, res.StatusCode, nil
}
