package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

var usage = `rerun runs the passed command for each line in standard input.
Previous invocations of the command are cancelled.

Usage:
	rerun [command]

Example:
	chokidar . | rerun go run main.go
`

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("must provide command to run")
	}

	cmdName, args := func() (name string, args []string) {
		name = flag.Arg(0)
		if flag.NArg() > 1 {
			args = flag.Args()[1:]
		}
		return
	}()

	restart := make(chan error, 0)
	go func() {
		stdin := bufio.NewReader(os.Stdin)
		for {
			_, err := stdin.ReadString('\n')
			restart <- err
		}
	}()

	for {
		ctx, cancel := context.WithCancel(context.Background())

		cmd := exec.CommandContext(ctx, cmdName, args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		log.Printf("running %q\n", cmd.Args)
		if err := cmd.Start(); err != nil {
			log.Println(err)
		}

		err := <-restart
		cancel()

		if err != nil {
			log.Println(err)
			return
		}
	}
}
