package metadata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testMetaData = `<?xml version="1.0" encoding="UTF-8"?>
<metadata modelVersion="1.1.0">
  <groupId>log4j</groupId>
  <artifactId>log4j</artifactId>
  <versioning>
    <latest>1.2.17</latest>
    <release>1.2.17</release>
    <versions>
      <version>1.1.3</version>
      <version>1.2.4</version>
      <version>1.2.5</version>
      <version>1.2.6</version>
      <version>1.2.7</version>
      <version>1.2.8</version>
      <version>1.2.9</version>
      <version>1.2.11</version>
      <version>1.2.12</version>
      <version>1.2.13</version>
      <version>1.2.14</version>
      <version>1.2.15</version>
      <version>1.2.16</version>
      <version>1.2.17</version>
    </versions>
    <lastUpdated>20140318154402</lastUpdated>
  </versioning>
</metadata>`
	testMetaDataSHA1 = `d290cc8eba0504881f1d165820c27fd7ea5b1d0f`
)

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.RequestURI, ".sha1") {
			fmt.Fprintln(w, testMetaDataSHA1)
		}
		fmt.Fprintln(w, testMetaData)
	}))
	defer ts.Close()
	md, err := Get(ts.URL, "log4j", "log4j", http.DefaultClient)
	if err != nil {
		t.Fatalf("Error getting metadata: %v", err)
	}
	assert.Equal(t, "1.2.17", md.Versioning.Latest)
	t.Logf("%+v", md)
}
