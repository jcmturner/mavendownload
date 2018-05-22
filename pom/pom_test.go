package pom

import (
	"testing"

	"github.com/jcmturner/mavendownload/metadata"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	md, err := metadata.Get("http://central.maven.org/maven2", "log4j", "log4j")
	if err != nil {
		t.Fatalf("error getting repo metadata: %v", err)
	}
	p, err := Get("http://central.maven.org/maven2", "log4j", "log4j", md.Versioning.Latest)
	if err != nil {
		t.Fatalf("error getting pom: %v", err)
	}
	assert.Equal(t, "log4j", p.GroupID, "GroupID not as expected")
	assert.Equal(t, "log4j", p.ArtifactID, "ArtifactID not as expected")
}
