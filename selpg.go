package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type Selpg_args struct {
	start_page          int
	end_page            int
	input_filename      string
	dest                string
	page_len            int
	form_feed_delimited bool
}

var progname string

func main() {
	progname = os.Args[0]

	var args Selpg_args

	FlagInit(&args)

	Process_args(&args)
	Process_input(&args)
}

func FlagInit(args *Selpg_args) {
	flag.Usage = usage
	flag.IntVar(&args.start_page, "s", -1, "Start page.")
	flag.IntVar(&args.end_page, "e", -1, "End page.")
	flag.IntVar(&args.page_len, "l", 72, "Line number per page.")
	flag.BoolVar(&args.form_feed_delimited, "f", false,
		"Determine form-feed-delimited")
	flag.StringVar(&args.dest, "d", "", "Destination.")
	flag.Parse()
}

func usage() {
	fmt.Printf("Usage of %s:\n\n", progname)
	fmt.Printf("%s is a tool to select pages from what you want.\n\n",
		progname)
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\tselpg -s=Number -e=Number [options] [filename]\n\n")
	fmt.Printf("The arguments are:\n\n")
	fmt.Printf("\t-s=Number\tStart from Page <number>.\n")
	fmt.Printf("\t-e=Number\tEnd to Page <number>.\n")
	fmt.Printf("\t-l=Number\t[options]Specify the number of line per page.Default is 72.\n")
	fmt.Printf("\t-f\t\t[options]Specify that the pages are sperated by \\f.\n")
	fmt.Printf("\t-d\t\t[options]Pipe to another program.\n")
	fmt.Printf("\t[filename]\t[options]Read input from the file.\n\n")
	fmt.Printf("If no file specified, %s will read input from stdin. Control-D to end.\n\n", progname)
}

/* Some Checks. */
func Process_args(args *Selpg_args) {
	if args.start_page == -1 || args.end_page == -1 {
		fmt.Fprintf(os.Stderr, "%s: not enough arguments\n\n", progname)
		flag.Usage()
		os.Exit(1)
	}

	if os.Args[1][0] != '-' || os.Args[1][1] != 's' {
		fmt.Fprintf(os.Stderr, "%s: 1st arg should be -sstart_page\n\n", progname)
		flag.Usage()
		os.Exit(1)
	}

	end_index := 2
	if len(os.Args[1]) == 2 {
		end_index = 3
	}

	if os.Args[end_index][0] != '-' || os.Args[end_index][1] != 'e' {
		fmt.Fprintf(os.Stderr, "%s: 2st arg should be -eend_page\n\n", progname)
		flag.Usage()
		os.Exit(1)
	}

	if args.start_page > args.end_page || args.start_page < 0 ||
		args.end_page < 0 {
		fmt.Fprintln(os.Stderr, "Invalid arguments")
		flag.Usage()
		os.Exit(1)
	}
}

func Process_input(args *Selpg_args) {
	/* Pipe initialization. */
	var cmd *exec.Cmd
	var cmdIn io.WriteCloser
	var cmdOut io.ReadCloser
	if args.dest != "" {
		cmd = exec.Command("bash", "-c", args.dest)
		cmdIn, _ = cmd.StdinPipe()
		cmdOut, _ = cmd.StdoutPipe()
		cmd.Start()
	}

	if flag.NArg() > 0 {
		/* Input from file. */
		args.input_filename = flag.Arg(0)
		output, err := os.Open(args.input_filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		reader := bufio.NewReader(output)
		count := 0
		for {
			line, _, err := reader.ReadLine()
			if err != io.EOF && err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if err == io.EOF {
				break
			}
			if count/args.page_len >= args.start_page &&
				count/args.page_len <= args.end_page {
				if args.dest != "" {
					fmt.Fprintln(cmdIn, string(line))
				} else {
					fmt.Println(string(line))
				}
			}
			count++
		}
		if args.dest != "" {
			cmdIn.Close()
			cmdBytes, _ := ioutil.ReadAll(cmdOut)
			cmd.Wait()
			fmt.Print(string(cmdBytes))
		}
	} else {
		/* Input from screen. */
		scanner := bufio.NewScanner(os.Stdin)
		count := 0
		target := ""

		for scanner.Scan() {
			line := scanner.Text()
			line += "\n"
			if count/args.page_len >= args.start_page &&
				count/args.page_len <= args.end_page {
				target += line
			}
			count++
		}
		if args.dest != "" {
			fmt.Fprint(cmdIn, target)
			cmdIn.Close()
			cmdBytes, _ := ioutil.ReadAll(cmdOut)
			cmd.Wait()
			fmt.Print(string(cmdBytes))
		} else {
			fmt.Print(target)
		}

	}

}
