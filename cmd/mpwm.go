package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

// this software's implementation is heavily rely on
// [gotcpproxy](https://github.com/jokeofweek/gotcpproxy/).

func main() {
	app := cli.NewApp()
	app.Name = "mpwm - messagepack viewer"
	app.Usage = "show message pack valuew from file or http proxy"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "from, f",
			Value: "localhost:80",
			Usage: "The proxy server's host.",
		},
		cli.StringFlag{
			Name:  "to, t",
			Value: "localhost:8000",
			Usage: "The host that the proxy server should forward requests to.",
		},
	}

	app.Action = func(c *cli.Context) {
		from := c.String("from")
		fmt.Println(from)
		to := c.String("to")
		fmt.Println(to)
	}

	app.Run(os.Args)
}
