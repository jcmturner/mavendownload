package metadata

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	LastUpdatedLayout = "20060102150405"
	mavenMetadataFile = "maven-metadata.xml"
)

type MetaData struct {
	ModelVersion string     `xml:"modelVersion,attr"`
	GroupID      string     `xml:"groupId"`
	ArtifactID   string     `xml:"artifactId"`
	Versioning   Versioning `xml:"versioning"`
}

type Versioning struct {
	Latest         string    `xml:"latest"`
	Release        string    `xml:"release"`
	Versions       []string  `xml:"versions>version"`
	LastUpdatedStr string    `xml:"lastUpdated"`
	LastUpdated    time.Time `xml:"-"`
}

func Get(repo, groupID, artifactID string) (md MetaData, err error) {
	url := fmt.Sprintf("%s/%s/%s/%s", strings.TrimRight(repo, "/"), groupID, artifactID, mavenMetadataFile)

	// Get the metadata
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("error getting %s: %v", url, err)
		return
	}
	cl := http.DefaultClient
	resp, err := cl.Do(req)
	mb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading body from %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	mdsha1, err := SHA1(url)
	if err != nil {
		err = fmt.Errorf("error getting SHA1: %v", err)
		return
	}

	// Check the md5 of the metadata
	hash := sha1.New()
	hash.Write(mb)
	h := hex.EncodeToString(hash.Sum(nil))
	if h != mdsha1 {
		err = fmt.Errorf("checksum of metadata does not match. expected: %s got: %s", mdsha1, h)
		return
	}

	// Marshal bytes into MetaData type
	rdr := bytes.NewReader(mb)
	decoder := xml.NewDecoder(rdr)
	err = decoder.Decode(&md)
	if err != nil {
		err = fmt.Errorf("error decoding metadata from %s: %v", url, err)
		return
	}
	err = md.Versioning.parseLUpdate()
	return
}

func (v *Versioning) parseLUpdate() (err error) {
	if v.LastUpdatedStr == "" {
		return
	}
	v.LastUpdated, err = time.Parse(LastUpdatedLayout, v.LastUpdatedStr)
	return
}

func SHA1(url string) (string, error) {
	url = url + ".sha1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("%s: %v", url, err)
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	r := bufio.NewReader(resp.Body)
	mdsha1, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		err = fmt.Errorf("%s: %v", url, err)
		return "", err
	}
	return strings.TrimSpace(mdsha1), nil
}
