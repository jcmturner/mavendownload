package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jcmturner/mavendownload/download"
)

func main() {
	repo := flag.String("repo", "", "Maven repository root URL")
	grpID := flag.String("groupid", "", "GroupID of artifact")
	artifactID := flag.String("artifactid", "", "ArtifactID of artifact")
	ext := flag.String("ext", "", "Override file extension")
	out := flag.String("out", "", "Output path")
	flag.Parse()

	n, fname, err := download.Latest(*repo, *grpID, *artifactID, *ext, *out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Downloaded %d bytes to %s/%s", n, strings.TrimRight(*out, "/"), fname)
}
