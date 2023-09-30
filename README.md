# MetaErr

[![GoDoc](https://pkg.go.dev/badge/github.com/quantumcycle/metaerr)](https://pkg.go.dev/github.com/quantumcycle/metaerr?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/quantumcycle/metaerr)](https://goreportcard.com/report/github.com/quantumcycle/metaerr)
[![codecov](https://codecov.io/gh/quantumcycle/metaerr/graph/badge.svg?token=3EFILQUGE9)](https://codecov.io/gh/quantumcycle/metaerr)

Metaerr is a golang package to create or wrap errors with custom metadata and location.

## Why

I was using github.com/pkg/errors before and the stacktraces were huge and not really useful. Then I found the [Fault library](https://github.com/Southclaws/fault) which was amazing, but the usage I wanted to do was at odd with some of the opinions built into the library.

This is why I decided to write this simple library. It uses the same "stacktrace" models as `Fault`, in the sense that the stack you will see if the locations of the errors (and wrapped errors) creation, and not the stacktrace that led to the creation of the error.

And the next feature it provides is the ability to add any number of key/pair metadata entries to each error (and wrapped errors). This is useful if you want to attach metadata at error create and then leverage that metadata at resolution. A common use case is to have a generic http error handler for an API that would leverage the metadata to determine the http status or build an error payload to be sent to the user. Another use case would be logging and alerting. If you convert the metadata into fields in a JSON logger, you could have different alerting rules for logged ERRORS based on the metadata, for example, errors with the metadata `tag` containing "security" could raise a immediate alert.

## Install

```
go get -u github.com/quantumcycle/metaerr
```

## Usage

Metaerr can be used with golang standard errors package. They are also compatible with error wrapping introduced in Go 1.13.

To create an new MetaErr from a string, use

```golang
err := metaerr.New("failure")
```

To create a new MetaErr by wrapping an existing error, use

```golang
err := metaerr.Wrap(err, "failure")
```

The next step once you have a Metaerr is to add metadata to it. The first step is to create a function matching the `metaerr.ErrorMetadata` signature. For your convenience, you can use `metaerr.StringMeta`, but you can also create your own. Ultimately, all metadata entries are stored as string.

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

You can provide options to alter the errors on creation. At the moment there is a single option, `WithLocationSkip`. By default when creating an error, Metaerr will skip 2 call stack to determine the error creation location. This works well if you call metaerr directly at the place where the error is created. However there is a use case where you want to create an error factory for some common use case, to initialize the error with some common metadata. In this case, if you use the standard `metaerr.New`, the reported location will be the line where `metaerr.New` is called, which will be in your error factory method. This is probably not what you want. In this case, you can use the `metaerr.WithLocationSkip` option to add additional call stack skip to determine the location. Here is an example:

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
        at /home/matdurand/sources/github/quantumcycle/metaerr/cmd/main.go:16
```

Without the `WithLocationSkip` option, the reported location would be line 12, inside the `CreateDatabaseError` function. having all our errors pointing to this specific line would ne useless.
