package main

import (
	"net/http"
	"bufio"
	"os"
	"bytes"
	"encoding/json"
	"log"
	"flag"
	"os/user"
)

var cfg JsonConfig = JsonConfig{}
var settings = [2]bool{false, false} //quiet, ignore empty
var version = "1.0"

func main() {

	flag.BoolVar(&settings[0], "q", false, "Only output the final url")
	flag.BoolVar(&settings[1], "i", false, "Ignore empty file/stream")

	flag.Parse()

	fi, _ := os.Stdin.Stat()

	loadConfig();

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		pipe()
		return
	}

	if cap(os.Args) < 2 {
		println("Usage: haste <file>")
		os.Exit(1)
	}

	filename := os.Args[cap(os.Args) - 1]

	if _, err := os.Stat(filename); err == nil {

		file, _ := os.Open(filename)
		buf := bytes.NewBuffer(nil)

		buf.ReadFrom(file)

		if buf.Len() == 0 && !settings[1] {
			println("File is empty. Ignoring it.")
			return
		}

		post(*buf)

	} else {
		println("File not found or not accessible.")
		os.Exit(1)
	}

}

func loadConfig() {
	usr, err := user.Current()

	if err != nil {
		println("Unable to find user home directory")
		os.Exit(1)
	}

	filename := usr.HomeDir + "/.hasterc"

	if _, err := os.Stat(filename); err == nil { // File exists
		file, err := os.Open(filename)
		CheckError(err)

		buf := bytes.NewBuffer(nil)
		buf.ReadFrom(file)

		CheckError(json.Unmarshal(buf.Bytes(), &cfg));

		defer file.Close()
		return
	}

	file, err := os.Create(filename)
	CheckError(err)

	config := JsonConfig{ FrontendUrl: "https://hastebin.com", BackendUrl: "https://hastebin.com" }
	b, err := json.MarshalIndent(config, " ", "  ")
	CheckError(err)

	file.Write(b)
	println("# No config found. Created one: " + filename)

	cfg.BackendUrl = config.BackendUrl
	cfg.FrontendUrl = config.FrontendUrl

	defer file.Close()

}

func pipe() {
	reader := bufio.NewReader(os.Stdin)
	buf := bytes.NewBuffer(nil)

	buf.ReadFrom(reader)
	if buf.Len() <= 0 && !settings[1] {
		println("Input is empty. Ignoring it.")
		return
	}

	post(*buf)
}

func post(buf bytes.Buffer) {

	client := &http.Client{ Timeout: 5 * 1000000000 }


	req, err := http.NewRequest("POST", cfg.BackendUrl + "/documents", &buf)
	CheckError(err)

	req.Header.Set("User-Agent", "Go-Haste/" + version)

	resp, err := client.Do(req)
	CheckError(err)

	decoder := json.NewDecoder(resp.Body)

	var result HasteResult
	CheckError(decoder.Decode(&result))

	if len(result.Key) == 0 {
		println("Failed to create a haste")
		println("Reason: " + result.Message)
		return
	}

	if settings[0] {
		println(cfg.FrontendUrl + "/" + result.Key)
		return
	}

	println("Url: " + cfg.FrontendUrl + "/" + result.Key)

}

type JsonConfig struct {
	BackendUrl, FrontendUrl string
}

type HasteResult struct {
	Key, Message string
}

//noinspection GoDuplicate
func CheckError(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}
