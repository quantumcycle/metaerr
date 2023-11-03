# MetaErr

[![GoDoc](https://pkg.go.dev/badge/github.com/quantumcycle/metaerr)](https://pkg.go.dev/github.com/quantumcycle/metaerr?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/quantumcycle/metaerr)](https://goreportcard.com/report/github.com/quantumcycle/metaerr)
[![codecov](https://codecov.io/gh/quantumcycle/metaerr/graph/badge.svg?token=3EFILQUGE9)](https://codecov.io/gh/quantumcycle/metaerr)

Metaerr is a golang package to create or wrap errors with custom metadata and location.

## Why

I used github.com/pkg/errors before, and the stack traces were extensive (like Java) and not very useful. Then, I came across the [Fault library](https://github.com/Southclaws/fault), which was amazing, but the way I wanted to use it clashed with some of the opinions embedded in the library.

This is why I decided to create this simple library. It utilizes the same "stack trace" model as Fault, in the sense that you will see the stack pertaining to the locations of error creation, and not the stack trace that led to the error's creation.

The next feature it offers is the ability to add any number of key-value metadata entries to each error, including wrapped errors. This is useful if you want to attach metadata at the time of error creation and then leverage that metadata during resolution. A common use case is having a generic HTTP error handler for an API that can use the metadata to determine the HTTP status or construct an error payload to send to the user. Another use case would be logging and alerting. If you convert the metadata into fields in a JSON logger, you could have different alerting rules for logged ERRORS based on the metadata; for example, errors with the metadata tag containing "security" could trigger an immediate alert.

## Install

```
go get -u github.com/quantumcycle/metaerr
```

## Usage

Metaerr can be used with the Go standard errors package, and they are also compatible with error wrapping introduced in Go 1.13.

To create an new MetaErr from a string, use

```golang
err := metaerr.New("failure")
```

To create a new MetaErr by wrapping an existing error, use

```golang
err := metaerr.Wrap(err, "failure")
```

The next step, once you have a Metaerr, is to add metadata to it. You need to create a function that matches the `metaerr.ErrorMetadata` signature. For your convenience, you can use `metaerr.StringMeta`, but you can also create your own. Ultimately, all metadata entries are stored as strings.

```golang
//Create an metadata called ErrorCode
var ErrorCode = metaerr.StringMeta("error_code")

func main() {
	rootCause := metaerr.New("failure")
	err := metaerr.Wrap(rootCause, "cannot fetch content").Meta(ErrorCode("x01"))
	fmt.Printf("%+v", err)
}
```

will print (... will be your project location)

```
cannot fetch content [error_code=x01]
        at .../quantumcycle/metaerr/cmd/main.go:13
failure
        at .../quantumcycle/metaerr/cmd/main.go:12
```

### How to use this library in your project

You can start with the example above, but if for example you want stacktraces, having this code everywhere is not really convenient

```golang
var ErrorCode = metaerr.StringMeta("error_code")

...
...
...

err := metaerr.New("failure", metaerr.WithStackTrace(0, 3)).Meta(ErrorCode("x01"))
```

To solve this, metaerr provides a base builder to help build your own error builder. You should create your own error builder in your project, adding any potential configuration options and metadata to the builder and then levering this to create your errors. 
Look at [this file](./example/errors/builder.go) for an example of a builder with 2 possible metadata (errorCode and tags), and [this file](./example/main.go) as an example of using this builder.

### Getting the err message, location, and metadata

In the example above, we use the Printf formatting to display the error, metadata and location all in one gulp. You can however use the provided helper function to get the individual parts

```golang
err := metaerr.New("failure")
err.Error() //returns failure
err.Location() //returns .../mysource/mypackage/file.go:22

// will print error_code:x01
meta := metaerr.GetMeta(err, false)
for k, values := range meta {
  for _, val := range values {
    fmt.Println(k + ":" + val)
  }
}

```

### Options

You can provide options to modify the errors during creation. 

#### WithLocationSkip
By default, when creating an error, Metaerr will skip 2 call stack frames to determine the error's creation location. This works well when you call Metaerr directly at the place where the error is created in your codebase. However, there is a use case where you might want to create an error factory function for common scenarios to initialize the error with some standard metadata. In this case, if you use the standard `metaerr.New` function, the reported location will be the line where `metaerr.New` is called, which may be within your error factory function. You probably don't want to have all your locations pointing to the same line. To address this, you can use the `metaerr.WithLocationSkip` option to add additional call stack skips to determine the location. Here is an example:

```golang
package main

import (
	"fmt"

	"github.com/quantumcycle/metaerr"
)

var Tag = metaerr.StringMeta("tag")

func CreateDatabaseError(reason string) error {
	return metaerr.New(reason, metaerr.WithLocationSkip(1)).Meta(Tag("database"))
}

func main() {
	dbErr := CreateDatabaseError("no such table [User]")
	fmt.Printf("%+v", dbErr)
}

```

which will output

```
no such table [User] [tag=database]
        at .../github.com/quantumcycle/metaerr/cmd/main.go:16
```

Without the `WithLocationSkip` option, the reported location would be line 12, inside the `CreateDatabaseError` function. Having all our errors pointing to this specific line would ne useless.

#### WithStacktrace

Usually the error creation location is enough to get by and find the context during which the error was created, but if the error is created in some central location called from multiple places, it might be useful to have a stacktrace to be able to find the caller that led to the error creation.
For these cases, use `WithStacktrace`, either when creating the error or when wrapping an existing error. When doing so, it will print something like this
```
failure
	at .../github.com/quantumcycle/metaerr/errors_test.go:46
	Stacktrace:
		.../github.com/quantumcycle/metaerr/errors_test.go:64
		.../github.com/quantumcycle/metaerr/errors_test.go:297
```

If you explore the unit test file, you will notice that the first line of the stacktrace is actually the error creation location, in this case, line 46 of the errors_test.go, but instead of repeating the same line twice, once for location and once at the start of the stacktrace, we skip it in the stacktrace.