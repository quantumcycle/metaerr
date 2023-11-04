# MetaErr

[![GoDoc](https://pkg.go.dev/badge/github.com/quantumcycle/metaerr)](https://pkg.go.dev/github.com/quantumcycle/metaerr?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/quantumcycle/metaerr)](https://goreportcard.com/report/github.com/quantumcycle/metaerr)
[![codecov](https://codecov.io/gh/quantumcycle/metaerr/graph/badge.svg?token=3EFILQUGE9)](https://codecov.io/gh/quantumcycle/metaerr)

Metaerr is a golang package to create or wrap errors with custom metadata and location.

This library requires Golang 1.21+.

## Why

I used github.com/pkg/errors before, and the stack traces were extensive (like Java) and not very useful. Then, I came across the [Fault library](https://github.com/Southclaws/fault), which was amazing, but the way I wanted to use it clashed with some of the opinions embedded in the library.
There is also [samber oops library](https://github.com/samber/oops), but the problem was that it's not extendable to have custom metadata.

This is why I decided to create this simple library. It utilizes the same "stack trace" model as Fault, in the sense that you will see the stack pertaining to the locations of error creation, but also adds regular stacktraces on top of that as an option.

The next feature it offers is the ability to add any number of key-value metadata entries to each error, including wrapped errors. This is useful if you want to attach metadata at the time of error creation and then leverage that metadata during resolution. A common use case is having a generic HTTP error handler for an API that can use the metadata to determine the HTTP status or construct an error payload to send to the user. Another use case would be logging and alerting. If you convert the metadata into fields in a JSON logger, you could have different alerting rules for logged ERRORS based on the metadata; for example, errors with the metadata tag containing "security" could trigger an immediate alert.

## Install

```
go get -u github.com/quantumcycle/metaerr
```

## Usage

Metaerr can be used with the Go standard errors package, and they are also compatible with error wrapping introduced in Go 1.13.

There is 2 ways to use the library. Using the errors directly, or using the builder. **The builder approach is recommended.**

### using the errors directly

To create an new MetaErr from a string, use

```golang
err := metaerr.New("failure")
```

To create a new MetaErr by wrapping an existing error, use

```golang
err := metaerr.Wrap(err, "failure")
```

The if you want to add metadata, you first create the metadata, and pass it as an option.

```golang
//Create an metadata called ErrorCode
var ErrorCode = metaerr.StringMeta("error_code")

func main() {
	rootCause := metaerr.New("failure", metaerr.WithMeta(ErrorCode("x01"))
	err := metaerr.Wrap(rootCause, "cannot fetch content")
	fmt.Printf("%+v", err)
}
```
will print
```
cannot fetch content
        at .../quantumcycle/metaerr/cmd/main.go:12
failure [error_code=x01]
        at .../quantumcycle/metaerr/cmd/main.go:11

```

### using the builder

Using the errors directly is ok, but it's a bit verbose just to create errors. Using the builder is a better approach 
and reduce boilerplate.

To use the builder, just create an instance of the builder with the relevant options for you, and then use it.
```golang
package main

import (
	"fmt"
	"github.com/quantumcycle/metaerr"
)

var errors = metaerr.NewBuilder(metaerr.WithStackTrace(0, 2))
var ErrorCode = metaerr.StringMeta("error_code")

func main() {
	err := errors.Meta(ErrorCode("test")).Newf("failure with user %s", "test")
	fmt.Printf("%+v\n", err)
}
```

### creating your own customized builder

The builder this library provides can be use as a standalone builder, but you should consider creating your own builder
by decorating the provided builder. The reason is that it's still very verbose to pass each instance of the metadata
when creating an error.

Look at [this file](./example/errors/builder.go) for an example of a builder with 2 possible metadata (errorCode and 
tags), and [this file](./example/main.go) as an example of using this builder.

The builder is immutable/thread safe, so you can have a base builder and then call `.Context(ctx)` on it without impacting the
rest of your code using the same builder. It does share the same metadata slice though, but there is no way to modify
the slice after creation, so it's safe.

### Getting the err message, location, and metadata

In the example above, we use the Printf formatting to display the error, metadata and location all in one gulp. 
You can however use the provided helper function to get the individual parts

```golang
err := metaerr.New("failure")
err.Error() //returns failure
merr := metaerr.AsMetaErr(err)
merr.Location() //returns .../mysource/mypackage/file.go:22

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

#### WithMeta

This is the main option you would be using. It allows you to add metadata to the error. You can add as many metadata as you want.
The library propose 4 built-in metadata builders:
- StringMeta: to add a string metadata
- StringsMeta: to add a slice of string metadata
- StringerMeta: to add any type that implements the Stringer interface as metadata
- StringMetaFromContext: to add a string metadata from a context (see `WithContext` below)

#### WithLocationSkip
By default, when creating an error, Metaerr will skip all stack frames related to metaerr to determine the error's creation location. 
This works well when you call Metaerr directly at the place where the error is created in your codebase. However, there is a use case 
where you use a factory, or your own builder to create errors. In this case, if you use the standard `metaerr.New` function, the reported 
location will be the line where metaerr is called to create the error, which may be within your error factory or builder function. 
You probably don't want to have all your locations pointing to the same line. To address this, you can use the `metaerr.WithLocationSkip` 
option to add additional call stack skips to determine the location. Here is an example:

```golang
package main

import (
	"fmt"

	"github.com/quantumcycle/metaerr"
)

var Tag = metaerr.StringMeta("tag")

func CreateDatabaseError(reason string) error {
	return metaerr.New(reason, metaerr.WithLocationSkip(1), metaerr.WithMeta(Tag("database")))
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

Without the `WithLocationSkip` option, the reported location would be line 12, inside the `CreateDatabaseError` function. 
Having all our errors pointing to this specific line would ne useless.

#### WithStacktrace

Usually the error creation location is enough to get by and find the context during which the error was created, but if 
the error is created in some central location called from multiple places, it might be useful to have a stacktrace to be 
able to find the caller that led to the error creation.
For these cases, use `WithStacktrace`, either when creating the error or when wrapping an existing error. When doing so, 
it will print something like this
```
failure
	at .../github.com/quantumcycle/metaerr/errors_test.go:46   //<-- this is the error default location
    at .../github.com/quantumcycle/metaerr/errors_test.go:64   //<-- this is added by the WithStacktrace option
    at .../github.com/quantumcycle/metaerr/errors_test.go:297  //<-- this is added by the WithStacktrace option
```

#### WithContext

This option allows you to attach a context to the error. Then you can use `StringMetaFromContext` to retrieve data from
the context and set some metadata. This is useful if for example you have a user in your context and want to add user
information to each error.