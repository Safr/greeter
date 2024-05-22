package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"text/template"
)

type config struct {
	numTimes       int
	name           string
	outputHTMLPath string
}

var errInvalidPosArgSpecified = errors.New("More than one positional argument specified")

func createHTMLGreeter(c config) error {
	f, err := os.Create(c.outputHTMLPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	tmpl, err := template.New("greeterHtml").Parse("<h1>Hello {{.}}</h1>")
	if err != nil {
		return err
	}
	return tmpl.Execute(f, c.name)
}

func getName(r io.Reader, w io.Writer) (string, error) {
	scanner := bufio.NewScanner(r)
	msg := "Your name please? Press the Enter key when done.\n"
	fmt.Fprint(w, msg)

	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	name := scanner.Text()
	if len(name) == 0 {
		return "", errors.New("You didn't enter your name")
	}
	return name, nil
}

func greetUser(c config, w io.Writer) {
	msg := fmt.Sprintf("Nice to meet you %s\n", c.name)
	for i := 0; i < c.numTimes; i++ {
		fmt.Fprint(w, msg)
	}
}

func runCmd(r io.Reader, w io.Writer, c config) error {
	var err error
	if len(c.name) == 0 {
		c.name, err = getName(r, w)
		if err != nil {
			return err
		}
	}

	if len(c.outputHTMLPath) != 0 {
		return createHTMLGreeter(c)
	}

	greetUser(c, w)
	return nil
}

func validateArgs(c config) error {
	if !(c.numTimes > 0) {
		return errors.New("Must specify a number greater than 0")
	}

	if !(c.numTimes > 0) && len(c.outputHTMLPath) > 0 {
		return errors.New("Must specify a number greater than 0")
	}

	return nil
}

func parseArgs(w io.Writer, args []string) (config, error) {
	c := config{}
	fs := flag.NewFlagSet("greeter", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.Usage = func() {
		usageString := `
A greeter application which prints the name you entered a specified number of times.

Usage of %s: <options> [name]`
		fmt.Fprintf(w, usageString, fs.Name())
		fmt.Fprintln(w)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options: ")
		fs.PrintDefaults()
	}
	fs.IntVar(&c.numTimes, "n", 0, "Number of times to greet")
	fs.StringVar(&c.outputHTMLPath, "o", "", "Path to the HTML page containing the greeting")
	err := fs.Parse(args)
	if err != nil {
		return c, err
	}

	if fs.NArg() > 1 {
		return c, errInvalidPosArgSpecified
	}
	if fs.NArg() == 1 {
		c.name = fs.Arg(0)
	}
	return c, nil
}

func main() {
	c, err := parseArgs(os.Stderr, os.Args[1:])
	if err != nil {
		if errors.Is(err, errInvalidPosArgSpecified) {
			fmt.Fprintln(os.Stdout, err)
		}
		os.Exit(1)
	}
	err = validateArgs(c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
	err = runCmd(os.Stdin, os.Stdout, c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
