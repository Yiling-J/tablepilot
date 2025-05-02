package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("rustc", "-Vv")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "host:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				fmt.Print(fields[1])
				return
			}
		}
	}
}
