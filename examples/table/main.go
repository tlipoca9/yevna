package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-viper/mapstructure/v2"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				fmt.Printf("panic: %+v\n", err)
				os.Exit(1)
			}
			panic(r)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var res []map[string]any
	err := yevna.Command(ctx, "echo", `
Permissions Size User       Date Modified Name
.rw-r--r--@ 139M foobarbazq 21 Mar 16:44  ca.txt
drwxr-xr-x     - foobarbazq 21 Mar 16:44  cmd
drwxr-xr-x     - foobarbazq 21 Mar 09:42  cmdx
drwxr-xr-x     - foobarbazq 21 Mar 17:34  examples
drwxr-xr-x     - foobarbazq 21 Mar 17:36  execx
.rw-r--r--  1.2k foobarbazq 21 Mar 17:29  go.mod
.rw-r--r--   14k foobarbazq 21 Mar 17:29  go.sum
.rw-r--r--   220 foobarbazq 21 Mar 15:51  Makefile
drwxr-xr-x     - foobarbazq 21 Mar 17:29  parser
.rw-r--r--  4.8k foobarbazq 21 Mar 17:22  yevna.go`[1:]).
		Quiet().
		RunWithParser(parser.Table(), &mapstructure.DecoderConfig{Result: &res})
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	// Output:
	// [map[Date Modified:21 Mar 16:44 Name:ca.txt Permissions:.rw-r--r--@ Size:139M User:foobarbazq] map[Date Modified:21 Mar 16:44 Name:cmd Permissions:drwxr-xr-x Size:- User:foobarbazq] map[Date Modified:21 Mar 09:42 Name:cmdx Permissions:drwxr-xr-x Size:- User:foobarbazq] map[Date Modified:21 Mar 17:34 Name:examples Permissions:drwxr-xr-x Size:- User:foobarbazq] map[Date Modified:21 Mar 17:36 Name:execx Permissions:drwxr-xr-x Size:- User:foobarbazq] map[Date Modified:21 Mar 17:29 Name:go.mod Permissions:.rw-r--r-- Size:1.2k User:foobarbazq] map[Date Modified:21 Mar 17:29 Name:go.sum Permissions:.rw-r--r-- Size:14k User:foobarbazq] map[Date Modified:21 Mar 15:51 Name:Makefile Permissions:.rw-r--r-- Size:220 User:foobarbazq] map[Date Modified:21 Mar 17:29 Name:parser Permissions:drwxr-xr-x Size:- User:foobarbazq] map[Date Modified:21 Mar 17:22 Name:yevna.go Permissions:.rw-r--r-- Size:4.8k User:foobarbazq]]
}
