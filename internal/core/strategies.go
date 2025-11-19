package core

import (
	"bytes"
	"context"
	"encoding/json"
	"text/template"
)

// StaticStrategy returns a fixed JSON response.
type StaticStrategy struct {
	Response interface{}
}

func (s *StaticStrategy) Generate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return s.Response, nil
}

// TemplateStrategy uses Go templates to generate a response.
type TemplateStrategy struct {
	TemplateStr string
}

func (s *TemplateStrategy) Generate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	tmpl, err := template.New("response").Parse(s.TemplateStr)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		// If not JSON, return string
		return buf.String(), nil
	}
	return result, nil
}
