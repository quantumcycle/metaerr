package main

import (
	"context"
	"fmt"
	"github.com/quantumcycle/metaerr/example/errors"
)

func fn1(ctx context.Context) error {
	return errors.New().
		ErrorCode("x01").
		Tags("security", "db").
		Errorf(ctx, "cannot fetch user with id %s", "123")
}

func main() {
	ctx := context.WithValue(context.Background(), "user_id", "333444")
	fmt.Printf("%+v\n", fn1(ctx))
}

/***
	This will output
	cannot fetch user with id 123 [error_code=x01] [tag=db,security] [user=333444]
        at .../github.com/quantumcycle/metaerr/example/main.go:13
        at .../github.com/quantumcycle/metaerr/example/main.go:18

*/
