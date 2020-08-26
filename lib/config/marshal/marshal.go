package marshal

import (
	"encoding/json"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
)

type JsonEncoder struct{}

func (j *JsonEncoder) Marshal(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}
func (j *JsonEncoder) Unmarshal(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}
func (j *JsonEncoder) Extension() string {
	return "json"
}

type TomlEncoder struct{}

func (j *TomlEncoder) Marshal(value interface{}) ([]byte, error) {
	return toml.Marshal(value)
}
func (j *TomlEncoder) Unmarshal(data []byte, value interface{}) error {
	return toml.Unmarshal(data, value)
}
func (j *TomlEncoder) Extension() string {
	return "toml"
}

type YamlEncoder struct{}

func (j *YamlEncoder) Marshal(value interface{}) ([]byte, error) {
	return yaml.Marshal(value)
}
func (j *YamlEncoder) Unmarshal(data []byte, value interface{}) error {
	return yaml.Unmarshal(data, value)
}
func (j *YamlEncoder) Extension() string {
	return "yaml"
}
