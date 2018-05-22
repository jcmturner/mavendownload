package download

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/jcmturner/mavendownload/metadata"
	"github.com/jcmturner/mavendownload/pom"
)

// Latest writes the latest version of the artifact to the provided Writer.
// To override the file extension in the POM file provide it to the fileExt argument, otherwise pass "" for this value.
func Latest(repo, groupID, artifactID, fileExt, path string) (int64, string, error) {
	md, err := metadata.Get(repo, groupID, artifactID)
	if err != nil {
		return 0, "", err
	}
	p, err := pom.Get(repo, groupID, artifactID, md.Versioning.Latest)
	if err != nil {
		return 0, "", err
	}
	if fileExt == "" {
		fileExt = p.Packaging
	}
	fname := fmt.Sprintf("%s-%s.%s", artifactID, md.Versioning.Latest, fileExt)
	url := fmt.Sprintf("%s/%s/%s/%s/%s", strings.TrimRight(repo, "/"),
		groupID, artifactID, md.Versioning.Latest, fname)
	w, err := os.Create(fmt.Sprintf("%s/%s", strings.TrimRight(path, "/"), fname))
	if err != nil {
		return 0, "", err
	}
	n, err := Get(url, w)
	return n, fname, err
}

// Get downloads the target URL to the provided Writer. As part of the download the SHA1 is validated.
func Get(url string, w io.Writer) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	h, err := metadata.SHA1(url)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)
	n, err := io.Copy(w, tee)
	if err != nil {
		return n, err
	}
	hash := sha1.New()
	hash.Write(buf.Bytes())
	if hex.EncodeToString(hash.Sum(nil)) != h {
		return n, errors.New("checksum of download does not match")
	}
	return n, nil
}
