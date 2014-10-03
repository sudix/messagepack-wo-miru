package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/codegangsta/cli"
)

// this software's implementation is heavily rely on
// [gotcpproxy](https://github.com/jokeofweek/gotcpproxy/).

var (
	fromHost              string
	toHost                string
	maxConnections        int
	maxWaitingConnections int
)

func setAppInfo(app *cli.App) {
	app.Name = "mpwm - messagepack viewer"
	app.Usage = "show message pack valuew from file or http proxy"
	app.Version = "0.0.1"
}

func setFlags(app *cli.App) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "from, f",
			Value: "localhost:8080",
			Usage: "The proxy server's host.",
		},
		cli.StringFlag{
			Name:  "to, t",
			Value: "localhost:8090",
			Usage: "The host that the proxy server" +
				" should forward requests to.",
		},
		cli.IntFlag{
			Name:  "max-connection, c",
			Value: 25,
			Usage: "The maximum number of active " +
				"connection at any given time.",
		},
		cli.IntFlag{
			Name:  "max-wait-connection, wc",
			Value: 10000,
			Usage: "The maximum number of " +
				"connections that can be waiting to be served.",
		},
	}

}

func matchConnections(waiting chan net.Conn, spaces chan bool) {
	// Iterate over each connection in the waiting channel
	for connection := range waiting {
		// Block until we have a space.
		<-spaces
		// Create a new goroutine which will call the connection handler and
		// then free up the space.
		go func(connection net.Conn) {
			handleConnection(connection)
			spaces <- true
			log.Printf("Closed connection from %s.\n", connection.RemoteAddr())
		}(connection)
	}
}

func handleConnection(connection net.Conn) {
	//Always close our connection.
	defer connection.Close()

	// Try to connect to remote server.
	remote, err := net.Dial("tcp", toHost)
	if err != nil {
		// Exit out when an error occurs
		log.Print(err)
		return
	}
	defer remote.Close()

	// Create our channel which for completation, and our two chennels to
	// signal that a goroutine is done.
	complete := make(chan bool, 2)
	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)
	go copyContent("---->", connection, remote, complete, ch1, ch2)
	go copyContent("<----", remote, connection, complete, ch2, ch1)
	// Block until we've completed both goroutines!'
	<-complete
	<-complete
}

func copyContent(dest string, from, to net.Conn, complete, done, otherDone chan bool) {
	fmt.Println(dest)
	var err error = nil
	var bytes []byte = make([]byte, 256)
	var read int = 0
	for {
		select {
		// If we received a done message from the othergoroutine, we exit.
		case <-otherDone:
			complete <- true
			return
		default:

			// Read data from the source connection.
			from.SetReadDeadline(time.Now().Add(time.Second * 5))
			read, err = from.Read(bytes)
			fmt.Printf("%s", bytes)
			// If any errors occured, write to complete as we are done
			// (one of the connections closed).
			if err != nil {
				complete <- true
				done <- true
				return
			}
			// write data to the destination.
			to.SetWriteDeadline(time.Now().Add(time.Second * 5))
			_, err = to.Write(bytes[:read])
			// Same error checking.
			if err != nil {
				complete <- true
				done <- true
				return
			}

		}

	}
}

func main() {
	app := cli.NewApp()
	setAppInfo(app)
	setFlags(app)

	app.Action = func(c *cli.Context) {
		fromHost = c.String("from")
		toHost = c.String("to")
		maxConnections = c.Int("max-connection")
		maxWaitingConnections = c.Int("max-wait-connection")
		fmt.Printf("Proxying %s->%s.\r\n", fromHost, toHost)

		// Set up our listening server
		server, err := net.Listen("tcp", fromHost)
		if err != nil {
			log.Fatal(err)
			return
		}

		// The channel of connections which are waiting to be processed.
		waiting := make(chan net.Conn, maxWaitingConnections)
		// The booleans representing the free active connection spaces.
		spaces := make(chan bool, maxConnections)
		// Initialize the spaces
		for i := 0; i < maxConnections; i++ {
			spaces <- true
		}

		// Start the connection matcher.
		go matchConnections(waiting, spaces)

		// Loop indefinitely, accepting connections and handling them.
		for {
			connection, err := server.Accept()
			if err != nil {
				// Log the error.
				log.Print(err)
				continue
			}

			// Create goroutine to handle the conn
			log.Printf("Received connection from %s.\n", connection.RemoteAddr())
			waiting <- connection
		}

	}

	app.Run(os.Args)
}
