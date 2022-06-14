package main

import (
	"bufio"
	"cx-im/src/config"
	"fmt"
	"github.com/moeshin/go-errs"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func buildKeyGo() {
	const name = "src/config/keys.go"
	file, err := os.Create(name)
	errs.Panic(err)
	defer errs.Close(file)
	_, err = file.WriteString("package config\nconst (\n")
	errs.Panic(err)
	for k := range config.KeyValues {
		_, err = file.WriteString(k)
		errs.Panic(err)
		_, err = file.WriteString("=\"")
		errs.Panic(err)
		_, err = file.WriteString(k)
		errs.Panic(err)
		_, err = file.WriteString("\"\n")
		errs.Panic(err)
	}
	_, err = file.WriteString(")")
	errs.Panic(err)
	errs.Panic(exec.Command("gofmt", "-w", name).Run())
}

func buildJsConfigValues(file *os.File) {
	regWord := regexp.MustCompile(`(\w+)`)
	iota := 0
	_, err := file.WriteString("class ConfigValues {\n")
	errs.Panic(err)
	f, err := os.Open("src/config/values.go")
	errs.Panic(err)
	defer errs.Close(f)
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	var start bool
	for s.Scan() {
		s := s.Text()
		s = strings.TrimSpace(s)
		if start {
			if s == "" || strings.HasPrefix(s, "//") {
				continue
			}
			if s == ")" {
				break
			}
			i := strings.IndexByte(s, ' ')
			if i != -1 {
				ss := strings.TrimSpace(s[i:])
				s = s[:i]
				if strings.HasPrefix(ss, "= ") {
					ss = ss[2:]
					_, err = file.WriteString(
						fmt.Sprintf("    static %s = %s;\n", s, regWord.ReplaceAllString(ss, "this.$1")))
					errs.Panic(err)
					continue
				}
			}
			_, err = file.WriteString(fmt.Sprintf("    static %s = %d;\n", s, 1<<iota))
			errs.Panic(err)
			iota++
		} else {
			if s == "const (" {
				start = true
			}
		}
	}
	_, err = file.WriteString("}\n")
	errs.Panic(err)
}

func buildJsConfigKeys(file *os.File) {
	_, err := file.WriteString("window.ConfigKey = {\n")
	errs.Panic(err)
	for k := range config.KeyValues {
		_, err = file.WriteString("    ")
		errs.Panic(err)
		_, err = file.WriteString(k)
		errs.Panic(err)
		_, err = file.WriteString(": \"")
		errs.Panic(err)
		_, err = file.WriteString(k)
		errs.Panic(err)
		_, err = file.WriteString("\",\n")
		errs.Panic(err)
	}
	_, err = file.WriteString("};\n")
	errs.Panic(err)
}

func buildJsConfigKeyValues(file *os.File) {
	_, err := file.WriteString("window.ConfigKeyValues = {\n")
	errs.Panic(err)
	for k, v := range config.KeyValues {
		_, err = file.WriteString("    ")
		errs.Panic(err)
		_, err = file.WriteString(k)
		errs.Panic(err)
		_, err = file.WriteString(": ")
		errs.Panic(err)
		_, err = file.WriteString(strconv.Itoa(v))
		errs.Panic(err)
		_, err = file.WriteString(",\n")
		errs.Panic(err)
	}
	_, err = file.WriteString("};\n")
	errs.Panic(err)
}

func buildJs() {
	const name = "web/assets/build.js"
	file, err := os.Create(name)
	errs.Panic(err)
	defer errs.Close(file)
	buildJsConfigValues(file)
	buildJsConfigKeys(file)
	buildJsConfigKeyValues(file)
}

func main() {
	buildKeyGo()
	buildJs()
}
