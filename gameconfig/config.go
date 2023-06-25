package gameconfig

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

//go:embed default.json
var defaultJSON []byte
var defaultProvider Provider

func init() {
	var err error
	defaultProvider, err = FromJSONStream(bytes.NewReader(defaultJSON))
	if err != nil {
		log.Fatalf("Error loading default gameconfig: %v - please report this.", err)
	}
}

type Provider interface {
	GetCharacterDefaults(id uint8) CharacterDefaults
	GetDefaultClubSetTypeID() uint32
}

type CharacterDefaults struct {
	CharacterID        uint8      `json:"CharacterID"`
	DefaultPartTypeIDs [24]uint32 `json:"DefaultPartTypeIDs"`
}

type Manifest struct {
	CharacterDefaults    []CharacterDefaults `json:"CharacterDefaults"`
	DefaultClubSetTypeID uint32              `json:"DefaultClubSetTypeID"`
}

type configFileProvider struct {
	characterDefaults    map[uint8]CharacterDefaults
	defaultClubSetTypeID uint32
}

func Default() Provider {
	return defaultProvider
}

func FromJSONFile(filename string) (Provider, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	provider, err := FromJSONStream(file)
	err = errors.Join(err, file.Close())
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func FromJSONStream(r io.Reader) (Provider, error) {
	manifest := Manifest{}
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	err := dec.Decode(&manifest)
	if err != nil {
		return nil, err
	}
	return FromManifest(manifest), nil
}

func FromManifest(manifest Manifest) Provider {
	provider := &configFileProvider{
		characterDefaults:    make(map[uint8]CharacterDefaults),
		defaultClubSetTypeID: manifest.DefaultClubSetTypeID,
	}
	for _, defaults := range manifest.CharacterDefaults {
		provider.characterDefaults[defaults.CharacterID] = defaults
	}
	return provider
}

func (c *configFileProvider) GetCharacterDefaults(id uint8) CharacterDefaults {
	return c.characterDefaults[id]
}

func (c *configFileProvider) GetDefaultClubSetTypeID() uint32 {
	return c.defaultClubSetTypeID
}
