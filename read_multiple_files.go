package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    //"io/ioutil"
    "os"
)

func main() {
    var filenames = [2]string{"file1.log", "file2.log"}
    var currentFile string

    for _, currentFile = range filenames {
        file := openFile(currentFile)
        defer file.Close()

        var reader io.Reader = file
        scanner := bufio.NewScanner(reader);

        for scanner.Scan() {
            line := scanner.Text();
            msg(line)
        }
    }
}

func openFile(filename string) *os.File {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    return file
}

func msg(output string) {
    fmt.Fprintln(os.Stderr, output)
}
