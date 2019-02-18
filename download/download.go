package download

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/jcmturner/mavendownload/metadata"
	"github.com/jcmturner/mavendownload/pom"
)

// Latest writes the latest version of the artifact into the specified path.
// To override the file extension in the POM file provide it to the fileExt argument, otherwise pass "" for this value.
func Latest(repo, groupID, artifactID, fileExt, path, caPath string) (int64, string, error) {
	cl, err := httpClient(caPath)
	if err != nil {
		return 0, "", err
	}
	md, err := metadata.Get(repo, groupID, artifactID, cl)
	if err != nil {
		return 0, "", err
	}
	return save(repo, groupID, artifactID, fileExt, path, md.Versioning.Latest, cl)
}

// Version writes the specified version of the artifact into the specified path.
// To override the file extension in the POM file provide it to the fileExt argument, otherwise pass "" for this value.
func Version(repo, groupID, artifactID, fileExt, version, path, caPath string) (int64, string, error) {
	cl, err := httpClient(caPath)
	if err != nil {
		return 0, "", err
	}
	md, err := metadata.Get(repo, groupID, artifactID, cl)
	if err != nil {
		return 0, "", err
	}
	var valid bool
	for _, v := range md.Versioning.Versions {
		if v == version {
			valid = true
			break
		}
	}
	if !valid {
		return 0, "", fmt.Errorf("version %s not available", version)
	}
	return save(repo, groupID, artifactID, fileExt, path, version, cl)
}

// Get downloads the target URL to the provided Writer. As part of the download the SHA1 is validated.
func Get(url string, w io.Writer, cl *http.Client) (int64, error) {
	if cl == nil {
		cl = http.DefaultClient
	}
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := cl.Do(r)
	if err != nil {
		return 0, err
	}
	h, err := metadata.SHA1(url, cl)
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

func save(repo, groupID, artifactID, fileExt, path, version string, cl *http.Client) (int64, string, error) {
	p, err := pom.Get(repo, groupID, artifactID, version, cl)
	if err != nil {
		return 0, "", err
	}
	if fileExt == "" {
		fileExt = p.Packaging
	}
	fname := fmt.Sprintf("%s-%s.%s", artifactID, version, fileExt)
	url := fmt.Sprintf("%s/%s/%s/%s/%s", strings.TrimRight(repo, "/"),
		groupID, artifactID, version, fname)
	w, err := os.Create(fmt.Sprintf("%s/%s", strings.TrimRight(path, "/"), fname))
	if err != nil {
		return 0, "", err
	}
	n, err := Get(url, w, cl)
	if err != nil {
		os.Remove(w.Name())
	}
	return n, fname, err
}

func httpClient(caPath string) (*http.Client, error) {
	if caPath == "" {
		return http.DefaultClient, nil
	}
	cp := x509.NewCertPool()
	// Load our trusted certificate path
	pemData, err := ioutil.ReadFile(caPath)
	if err != nil {
		return http.DefaultClient, err
	}
	ok := cp.AppendCertsFromPEM(pemData)
	if !ok {
		return http.DefaultClient, errors.New("could not append cert to cert pool trust store")
	}
	cl := http.DefaultClient
	if transport, ok := cl.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = &tls.Config{RootCAs: cp}
		cl.Transport = transport
		return cl, nil
	} else {
		tlsConfig := &tls.Config{RootCAs: cp}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		cl.Transport = transport
		return cl, nil
	}
}
