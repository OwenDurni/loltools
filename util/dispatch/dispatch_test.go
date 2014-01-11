package dispatch

import (
  "fmt"
  "net/http"
)

var paths = []string{
  "/",
  "/foo",
  "/foo/",
  "/foobar",
  "/foo/bar",
  "/foo/bar/baz",
  "/foo?arg=1",
  "/foo#arg=1",
}

func myHandler1(w http.ResponseWriter, r *http.Request, args map[string]string) {}

func printMatchingPaths(dispatcher *Dispatcher) {
  for _, path := range paths {
    handler, args, err := dispatcher.Select(path)
    if handler != nil {
      if len(args) == 0 {
        fmt.Println(path)
      } else {
        fmt.Println(path, args)
      }
    }
    if err != nil {
      fmt.Println("Error: %v", err)
    }
  }
}

func ExampleDispatcher_root() {
  dispatcher := new(Dispatcher)
  dispatcher.Add("/", myHandler1)
  printMatchingPaths(dispatcher)
  // Output:
  // /
}

func ExampleDispatcher_literal() {
  dispatcher := new(Dispatcher)
  dispatcher.Add("/foo", myHandler1)
  printMatchingPaths(dispatcher)
  // Output:
  // /foo
  // /foo/
  // /foo?arg=1
  // /foo#arg=1
}

func ExampleDispatcher_argument() {
  dispatcher := new(Dispatcher)
  dispatcher.Add("/foo/<arg>", myHandler1)
  printMatchingPaths(dispatcher)
  // Output:
  // /foo/bar map[arg:bar]
}

func ExampleDispatcher_multiargument() {
  dispatcher := new(Dispatcher)
  dispatcher.Add("/<first>/<second>", myHandler1)
  printMatchingPaths(dispatcher)
  // Output:
  // /foo/bar map[second:bar first:foo]
}

func ExampleDispatcher_star() {
  dispatcher := new(Dispatcher)
  dispatcher.Add("/foo/*", myHandler1)
  printMatchingPaths(dispatcher)
  // Output:
  // /foo/bar
  // /foo/bar/baz
}
