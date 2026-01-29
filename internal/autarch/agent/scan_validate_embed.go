package agent

import "embed"

//go:embed schemas/scan/*.json
var scanSchemasFS embed.FS

func SchemaFor(phase string) ([]byte, bool) {
	path := "schemas/scan/" + phase + ".json"
	data, err := scanSchemasFS.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func SynthesisSchema() []byte {
	data, _ := scanSchemasFS.ReadFile("schemas/scan/synthesis.json")
	return data
}
