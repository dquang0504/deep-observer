package validation

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Validator struct{
	EventSchema *jsonschema.Schema
	LogSchema	*jsonschema.Schema
}

func LoadSchemas(repoRoot string) (*Validator, error){
	eventPath := filepath.Join(repoRoot, "libs", "schemas", "event.schema.json")
	logPath := filepath.Join(repoRoot, "libs", "schemas", "log.schema.json")

	compiler := jsonschema.NewCompiler()
	// Allow loading schemas from local file paths.
	compiler.LoadURL = func(url string) (io.ReadCloser, error){
		//allow loading from file paths
		b, err := os.ReadFile(url)
		if err != nil{
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(b)), nil
	}

	if err := compiler.AddResource(eventPath, strings.NewReader(mustReadString(eventPath))); err != nil{
		return nil, fmt.Errorf("add event schema: %w", err)
	}
	if err := compiler.AddResource(logPath, strings.NewReader(mustReadString(logPath))); err != nil{
		return nil, fmt.Errorf("add log schema: %w", err)
	}

	ev, err := compiler.Compile(eventPath)
	if err != nil{
		return nil, fmt.Errorf("compile event schema: %w", err)
	}

	lg, err := compiler.Compile(logPath)
	if err != nil{
		return nil, fmt.Errorf("compile log schema: %w", err)
	}

	return &Validator{EventSchema: ev, LogSchema: lg}, nil
}

func mustReadString(path string) string{
	b, err := os.ReadFile(path)
	if err != nil{
		panic(err)
	}
	return string(b)
}
