package main

import (
	"context"
	"fmt"
	"github.com/quantumcycle/metaerr/example/errors"
)

func getProduct(ctx context.Context, id string) error {
	originalErr := fmt.Errorf("something went wrong")
	return errors.New().
		ErrorCode("x01").
		Tags("db").
		Wrap(ctx, originalErr, "cannot fetch product", "product_id", id)
}

func main() {
	ctx := context.WithValue(context.Background(), "user_id", "333444")
	err := getProduct(ctx, "9911")
	fmt.Printf("%+v\n", err)
}

/***
This will output:
--------------------------------------------------------------------------
cannot fetch product [error_code=x01] [product_id=9911] [tags=db] [user=333444]
	at .../github.com/quantumcycle/metaerr/example/main.go:14
	at .../github.com/quantumcycle/metaerr/example/main.go:19
something went wrong
--------------------------------------------------------------------------
error_code and tag comes at error creation time as first call builder metadata
product_id comes at error creating time as additional metadata attributes
user comes from the context attribute when present
*/
