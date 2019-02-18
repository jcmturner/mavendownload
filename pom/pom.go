package pom

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/jcmturner/mavendownload/metadata"
)

type Project struct {
	ModelVersion string       `xml:"modelVersion"`
	GroupID      string       `xml:"groupId"`
	ArtifactID   string       `xml:"artifactId"`
	Version      string       `xml:"version"`
	Packaging    string       `xml:"packaging"`
	Description  string       `xml:"description"`
	URL          string       `xml:"url"`
	Name         string       `xml:"name"`
	Licenses     []License    `xml:"licenses>license"`
	Dependencies []Dependency `xml:"dependencies>dependency"`
}

type License struct {
	Name         string `xml:"name"`
	URL          string `xml:"url"`
	Distribution string `xml:"distribution"`
}

type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Optional   bool   `xml:"optional"`
}

func Get(repo, groupID, artifactID, version string, cl *http.Client) (p Project, err error) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s-%s.pom", strings.TrimRight(repo, "/"), groupID, artifactID, version, artifactID, version)

	// Get the POM file
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("error forming request of %s: %v", url, err)
		return
	}
	if cl == nil {
		cl = http.DefaultClient
	}
	resp, err := cl.Do(req)
	if err != nil {
		err = fmt.Errorf("error getting %s: %v", url, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http response %d downloading POM file", resp.StatusCode)
		return
	}
	mb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading body from %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	// Get the POM file SHA1
	psha1, err := metadata.SHA1(url, cl)
	if err != nil {
		err = fmt.Errorf("error getting POM SHA1: %v", err)
		return
	}

	// Check the md5 of the metadata
	hash := sha1.New()
	hash.Write(mb)
	h := hex.EncodeToString(hash.Sum(nil))
	if h != psha1 {
		err = fmt.Errorf("checksum of POM does not match. expected: %s got: %s", psha1, h)
		return
	}

	// Marshal bytes into MetaData type
	rdr := bytes.NewReader(mb)
	decoder := xml.NewDecoder(rdr)
	err = decoder.Decode(&p)
	if err != nil {
		err = fmt.Errorf("error decoding POM from %s: %v", url, err)
		return
	}
	return
}
