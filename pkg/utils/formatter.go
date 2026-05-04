package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type OutputFormat string

const (
	FormatJSON    OutputFormat = "json"
	FormatText    OutputFormat = "text"
	FormatCompact OutputFormat = "compact"
)

type Formatter struct {
	format OutputFormat
	writer io.Writer
}

func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{
		format: format,
		writer: os.Stdout,
	}
}

func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

func (f *Formatter) Output(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.outputJSON(data)
	case FormatText:
		return f.outputText(data)
	case FormatCompact:
		return f.outputCompact(data)
	default:
		return f.outputJSON(data)
	}
}

func (f *Formatter) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data)
}

func (f *Formatter) outputText(data interface{}) error {
	switch v := data.(type) {
	case map[string]interface{}:
		return f.outputMapText(v)
	case []interface{}:
		return f.outputSliceText(v)
	default:
		_, err := fmt.Fprintf(f.writer, "%v\n", data)
		return err
	}
}

func (f *Formatter) outputMapText(data map[string]interface{}) error {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Fprintf(f.writer, "%s:\n", key)
			for k, val := range v {
				fmt.Fprintf(f.writer, "  %s: %v\n", k, val)
			}
		case []interface{}:
			fmt.Fprintf(f.writer, "%s:\n", key)
			for _, item := range v {
				fmt.Fprintf(f.writer, "  - %v\n", item)
			}
		default:
			fmt.Fprintf(f.writer, "%s: %v\n", key, value)
		}
	}
	return nil
}

func (f *Formatter) outputSliceText(data []interface{}) error {
	for i, item := range data {
		fmt.Fprintf(f.writer, "%d. %v\n", i+1, item)
	}
	return nil
}

func (f *Formatter) outputCompact(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data)
}

func OutputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data)
}

func OutputError(err error) error {
	return OutputJSON(map[string]interface{}{
		"status": "error",
		"error":  err.Error(),
	})
}

func OutputSuccess(data interface{}) error {
	return OutputJSON(map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

func OutputSuccessWithMeta(data interface{}, meta map[string]interface{}) error {
	return OutputJSON(map[string]interface{}{
		"status":   "success",
		"data":     data,
		"metadata": meta,
	})
}
