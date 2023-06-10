package iff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadItem(t *testing.T) {
	testLoadItemVersion(t, "testdata/eu301/item.iff")
	testLoadItemVersion(t, "testdata/eu500/item.iff")
	testLoadItemVersion(t, "testdata/in212/item.iff")
	testLoadItemVersion(t, "testdata/jp211/item.iff")
	testLoadItemVersion(t, "testdata/jp401/item.iff")
	testLoadItemVersion(t, "testdata/jp585/item.iff")
	testLoadItemVersion(t, "testdata/jp977/item.iff")
	testLoadItemVersion(t, "testdata/kr326/item.iff")
	testLoadItemVersion(t, "testdata/sg216/item.iff")
	testLoadItemVersion(t, "testdata/us431/item.iff")
	testLoadItemVersion(t, "testdata/us500/item.iff")
	testLoadItemVersion(t, "testdata/us852/item.iff")
}

func testLoadItemVersion(t *testing.T, filename string) {
	data := mustLoad(filename)
	file, err := LoadItems(data)
	if err != nil {
		t.Fatalf("Loading %q failed: %v", filename, err)
	}
	assert.Equal(t, "Test Item", file.Records[0].Name)
	assert.Equal(t, "item0_00", file.Records[0].Icon)
	assert.Equal(t, "item0_00", file.Records[0].Model)
}
