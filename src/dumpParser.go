package main

import (
	"./DataStructure"
	"./RevertCleaner"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func ParseDump(dumpFile string) {
	flag.Parse()

	cmd := exec.Command("7z", "x", dumpFile, "-so")
	out, _ := cmd.StdoutPipe()
	_ = cmd.Start()

	decoder := xml.NewDecoder(out)
	total := 0
	ignored := 0
	var inElement string
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			inElement = se.Name.Local
			if inElement == "page" {
				var p DataStructure.Page
				_ = decoder.DecodeElement(&p, &se)
				if p.Ns != "0" {
					ignored++
					continue
				}
				//WritePage(p)
				RevertCleaner.RevertBuilder(&p) // to make a pure parser, replace this line with Utils.WritePage("../out/", &p)
				total++
			}
		}
	}

	fmt.Printf("Total pages: %d \n", total)
	fmt.Printf("Total ignored pages: %d \n", ignored)
}

func main() {
	ParseDump(os.Args[1])
}
