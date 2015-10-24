package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"flag"
	"time"
)

func main() {
	job_flag := flag.Int("jobs", 0, "Show background jobs")
	sql_flag := flag.Int("sql", 0, "Show SQL statements")
	after_str := flag.String("after", "", "Show logs after this time (YYYY-MM-DD HH:II::SS")

	flag.Parse()
	args := flag.Args()

	// Time layouts must use the
	// reference time `Mon Jan 2 15:04:05 MST 2006` to show the
	// pattern with which to format/parse a given time/string
	time_layout      := "[2006-01-02 15:04:05 MST]"
	time_after, e    := time.Parse(time_layout, fmt.Sprintf("[%s UTC]", *after_str))
	parse_time       := false

	if e != nil {
		if len(*after_str) > 0 {
			fmt.Println(fmt.Sprintf("Invalid time format \"%s\" - Must be YYYY-MM-DD HH::II::SS", *after_str))
			os.Exit(4)
		}
	} else {
		parse_time = true
	}

	if len(args) < 2 {
		fmt.Println(fmt.Sprintf("Usage: avmlog -jobs=0|1 -sql=0|1 -after=\"YYYY-MM-DD HH:II::SS\" avmanager_filename.log regexp"))
		fmt.Println("Example: avm -jobs=1 \"/path/to/manager/log/production.log\" \"username|computername\"")
		os.Exit(1)
	}

	fmt.Println(fmt.Sprintf("Show background jobs: %d", *job_flag))
	fmt.Println(fmt.Sprintf("Show SQL: %d", *sql_flag))
	fmt.Println(fmt.Sprintf("After: %s", *after_str))

	filename := args[0]
	fmt.Println(fmt.Sprintf("Opening file: %s", filename))

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	line_count       := 0
	request_ids      := make([]string, 0)
	line_strexp      := args[1]

	line_regexp      := regexp.MustCompile(line_strexp) // ("apvuser03734|av-pd1-pl8-0787")
	timestamp_regexp := regexp.MustCompile("^(\\[.*?\\])")
	sql_regexp       := regexp.MustCompile("(SQL \\()|(EXEC sp_executesql N)|( CACHE \\()")
	nltm_regexp      := regexp.MustCompile(" \\(NTLM\\) ")
	target_regexp    := regexp.MustCompile("\\] (P[0-9]+[A-Za-z]+[0-9]+) ")

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text();
		if line_regexp.MatchString(line) {
			request   := target_regexp.FindStringSubmatch(line)
			timestamp := timestamp_regexp.FindStringSubmatch(line)

			input := true

			if len(request) < 2 {
				input = false
			} else if parse_time && len(timestamp) > 1 {
				line_time, e := time.Parse(time_layout, timestamp[1])

				if e != nil {
					fmt.Println("Got error %s", e)
					input = false
				} else if line_time.Before(time_after) {
					input = false
				}
			}

			if input {
				is_job := strings.Contains(request[1], "DJ")

				if is_job {
					if *job_flag > 0 {
						request_ids = append(request_ids, request[1])
					} else {
						fmt.Println(line)
					}
				} else {
					request_ids = append(request_ids, request[1])
				}
			}
		}

		line_count++

		if line_count % 10000 == 0 {
			fmt.Print(".")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generating unique request identifiers", len(request_ids))
	unique_set := make(map[string]bool, len(request_ids))

	for _, x := range request_ids {
		unique_set[x] = true
	}

	unique_ids := make([]string, 0, len(unique_set))

	for x := range unique_set {
		if len(x) > 0 {
			unique_ids = append(unique_ids, x)
		}
	}

	for i := 0; i < len(unique_ids); i++ {
		fmt.Println(fmt.Sprintf("Request ID: %s", unique_ids[i]))
	}

	unique_strexp := strings.Join(unique_ids, "|")
	fmt.Println(unique_strexp)

	if len(unique_strexp) < 1 {
		fmt.Println(fmt.Sprintf("Found 0 AVM Request IDs for %s", line_strexp))
		os.Exit(2)
	}

	file.Seek(0, 0)

	output_regexp := regexp.MustCompile(unique_strexp)
	output_scanner := bufio.NewScanner(file)

	for output_scanner.Scan() {
		line := output_scanner.Text();
		if output_regexp.MatchString(line) {

			output := true

			if *sql_flag < 1 && sql_regexp.MatchString(line) {
				output = false
			}

			if nltm_regexp.MatchString(line) {
				output = false
			}

			if output {
				fmt.Println(line)
			}
		}
	}

	if err := output_scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
