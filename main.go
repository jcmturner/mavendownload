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
	out := flag.String("out", "./", "Output path")
	flag.Parse()

	if !strings.HasPrefix(*repo, "http://") && !strings.HasPrefix(*repo, "https://") {
		fmt.Fprintln(os.Stderr, "Repo URL is invalid. Must start http:// or https://")
		os.Exit(1)
	}
	if *grpID == "" {
		fmt.Fprintln(os.Stderr, "GroupID is not defined")
		os.Exit(1)
	}
	if *artifactID == "" {
		fmt.Fprintln(os.Stderr, "ArtifactID is not defined")
		os.Exit(1)
	}
	n, fname, err := download.Latest(*repo, *grpID, *artifactID, *ext, *out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Output: %s/%s\nSize: %d", strings.TrimRight(*out, "/"), fname, n)
}
