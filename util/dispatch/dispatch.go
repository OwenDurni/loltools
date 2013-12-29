package dispatch

import (
  "errors"
  "fmt"
  "net/http"
  "strings"
)

type Handler func(http.Request, http.ResponseWriter, map[string]string)

type Dispatcher struct {
  root *dispatchNode
}

type dispatchNode struct {
  LiteralNexts  map[string]*dispatchNode
  ArgumentNexts map[string]*dispatchNode
  StarNext      *dispatchNode

  // The handler and pattern for this node if one exists.
  Handler Handler
  Pattern string
}

func (n *dispatchNode) init() *dispatchNode {
  n.LiteralNexts = make(map[string]*dispatchNode)
  n.ArgumentNexts = make(map[string]*dispatchNode)
  n.StarNext = nil
  return n
}

// Adds the specified handler with the specified pattern.
//
// The pattern format is described below.
//
// By default, the pattern is an exact-prefix match. Query parameters never
// matter for the purpose of dispatch.
//   PATTERN   : /foo
//   MATCHES   : /foo, /foo?arg=1
//   MISMATCHES: /foobar, /foo/bar
//
//   PATTERN   : /foo/bar
//   MATCHES   : /foo/bar, /foo/bar?arg=1
//   MISMATCHES: /foobar, /foo/bar/baz
//
// "/*" Can be used to construct a non-exact-match prefix. Note that paths like
// "/foo*" or "/foo/*/bar" are invalid. "/*" must terminate the pattern.
//   PATTERN   : /foo/*
//   MATCHES   : /foo/bar, /foo/bar/baz
//   MISMATCHES: /foo, /foobar
//
// "/<key>" Can be used as a wildcard for a path segment. The content of that path
// segment is forwarded to the dispatcher in the map.
//   PATTERN   : /foo/<a>
//   MATCHES   : /foo/1 (a:1), /foo/2 (a:2)
//   MISMATCHES: /foo, /foo/bar/baz, /foobar
//
//   PATTERN   : /foo/<a>/<b>
//   MATCHES   : /foo/1/2 (a:1,b:2)
//   MISMATCHES: /foo, /foo/1, /foo/1/2/3
func (dispatcher *Dispatcher) Add(pattern string, handler Handler) error {
  if dispatcher.root == nil {
    // Ensure root node is initialized.
    dispatcher.root = new(dispatchNode).init()
  }

  // Ensure pattern starts with "/"
  if !strings.HasPrefix(pattern, "/") {
    return errors.New(fmt.Sprintf("pattern must start with '/': %v", pattern))
  }

  // Split pattern into parts
  patternParts := strings.Split(pattern, "/")
  var node *dispatchNode = nil
  for i, patternPart := range patternParts {
    switch {
    case i == 0:
      node = dispatcher.root
    case len(patternPart) == 0:
      if i == len(patternParts)-1 {
        // Ignore trailing '/'
        break
      } else {
        return errors.New(fmt.Sprintf(
          "pattern cannot have two consecutive '/': %v", pattern))
      }
    case patternPart == "*":
      if i != len(patternParts)-1 {
        return errors.New(fmt.Sprintf(
          "pattern containing '*' must end with '*': %v", pattern))
      }
      if node.StarNext == nil {
        node.StarNext = new(dispatchNode).init()
      }
      node = node.StarNext
    case strings.HasPrefix(patternPart, "<"):
      if !strings.HasSuffix(patternPart, ">") {
        return errors.New(fmt.Sprintf(
          "pattern has unterminated '<' bracket: %v", pattern))
      } else {
        // TODO: This implementation allows conflicting patterns if the argument
        // name differs. Ex: Only one of '/<a>' and '/<b>' will be selected.
        argument := patternPart[1 : len(patternPart)-1]
        nextNode, exists := node.ArgumentNexts[argument]
        if !exists {
          nextNode = new(dispatchNode).init()
          node.ArgumentNexts[argument] = nextNode
        }
        node = nextNode
      }
    default:
      literal := patternPart
      nextNode, exists := node.LiteralNexts[literal]
      if !exists {
        nextNode = new(dispatchNode).init()
        node.LiteralNexts[literal] = nextNode
      }
      node = nextNode
    }
  }

  // If we got here then the pattern was valid and fully consumed.
  if node.Handler != nil {
    return errors.New(fmt.Sprintf(
      "pattern conflicts with existing pattern '%v': %v",
      node.Pattern, pattern))
  }
  node.Handler = handler
  node.Pattern = pattern
  return nil
}

func recSelect(
  pathIndex int,
  pathParts []string,
  node *dispatchNode,
  argAccum *map[string]string) Handler {
  // We do a DFS to match the most specific patterns first.

  // Base case: consumed the entire path, if there is a match return it.
  if pathIndex == len(pathParts) {
    return node.Handler
  }

  part := pathParts[pathIndex]

  // Base case: consumed the entire path, path has a trailing '/'. If there is
  // a match return it.
  if pathIndex == len(pathParts)-1 && part == "" {
    return node.Handler
  }

  // Recursive case: check for literal matches. Only return if there is a match.
  // Otherwise check next recursive case.
  nextNode, exists := node.LiteralNexts[part]
  if exists {
    handler := recSelect(pathIndex+1, pathParts, nextNode, argAccum)
    if handler != nil {
      return handler
    }
  }

  // Recursive case: check for argument segment matches. Only return if there is
  // a match. Otherwise check next recursive case.
  for arg, nextNode := range node.ArgumentNexts {
    handler := recSelect(pathIndex+1, pathParts, nextNode, argAccum)
    if handler != nil {
      (*argAccum)[arg] = part
      return handler
    }
  }

  // Recursive case: check for star matches.
  if node.StarNext != nil {
    return node.StarNext.Handler
  }

  // Recursive case: No matchers, return early.
  return nil
}

func (dispatcher *Dispatcher) Select(
  path string) (handler Handler, args map[string]string, err error) {
  handler, args, err = nil, make(map[string]string), nil

  // strip ?... and #...
  path = strings.Split(path, "?")[0]
  path = strings.Split(path, "#")[0]

  if !strings.HasPrefix(path, "/") {
    err = errors.New(fmt.Sprintf("input path is not absolute: %v", path))
    return
  }

  pathParts := strings.Split(path, "/")
  handler = recSelect(1, pathParts, dispatcher.root, &args)
  return
}

// Use http.HandleFunc("/", dispatcher.RootHandler) to use this dispatcher for
// dispatch in an application.
func (dispatcher *Dispatcher) RootHandler(r http.Request, w http.ResponseWriter) {

}
