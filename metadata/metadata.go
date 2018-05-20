package metadata

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	LastUpdatedLayout = "20060102150405"
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

func Get(repo string) (md MetaData, err error) {
	url := repo + "/maven-metadata.xml"
	urlSHA1 := url + ".sha1"

	// Get the metadata
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	cl := http.DefaultClient
	resp, err := cl.Do(req)
	mb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Get the metadata md5
	req, err = http.NewRequest("GET", urlSHA1, nil)
	if err != nil {
		return
	}
	respMD5, err := cl.Do(req)
	r := bufio.NewReader(respMD5.Body)
	mdsha1, err := r.ReadString('\n')
	mdsha1 = strings.TrimSpace(mdsha1)
	if err != nil {
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
