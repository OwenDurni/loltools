loltools
========

Collection of League of Legends tools for AppEngine.

Getting Started
---------------

See http://golang.org/doc/code.html for how to setup a canonical go workspace.

You can then initialize a git repository for this project with

```
go install github.com/OwenDurni/loltools
```

gofmt cmd
---------

```
gofmt -w -tabs=false -tabwidth=2 .
```

Application Directory Structure
-------------------------------

There are a few oddities about the directory structure for go appengine applications.
- Since SDK 1.7.4, the goapp command pulls dependencies from your $GOPATH and packages them with the deployed binary.
  (see http://blog.golang.org/the-app-engine-sdk-and-workspaces-gopath) goapp also sends all go files recursively
  under the application directory to the compiler. So if you have go files nested in subdirectories
  of the application root AND the application root is in your $GOPATH those files get included twice which results
  in an error.
- go encourages open source applications to live in a unique, fully-qualified location under the $GOPATH. For example,
  this project lives under $GOPATH/src/github.com/OwenDurni/loltools. The root of the git repository lives in the
  loltools directory. This means anything you want version controlled is in your $GOPATH. This means the application
  directory will be under your $GOPATH. Because of the first bullet point, this means application code must live
  outside the application root and must be imported with fully qualified path names.
- static files that need to be accessible from the app must be in subdirectories of the application root. So
  static resources and template files must be in subdirectories of the application root (and can contain no go files).
