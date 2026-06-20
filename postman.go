// Package postman implements a Postman Collection v2.1 exporter for Pucora API gateway
// configurations. It reads a Pucora service config and emits a Postman Collection JSON document.
//
// Extra config namespace: documentation/postman
package postman

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pucora/lura/v2/config"
	cmd "github.com/pucora/pucora-cobra/v2"
	"github.com/spf13/cobra"
)

// Namespace is the extra_config key used by this package.
const Namespace = "documentation/postman"

// EndpointConfig holds Postman-specific metadata that can be placed in an endpoint's
// extra_config under the "documentation/postman" namespace.
type EndpointConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// -- Postman Collection v2.1 structures --

type collectionInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type urlObject struct {
	Raw  string   `json:"raw"`
	Host []string `json:"host"`
	Port string   `json:"port"`
	Path []string `json:"path"`
}

type request struct {
	Method string    `json:"method"`
	URL    urlObject `json:"url"`
}

type item struct {
	Name    string  `json:"name"`
	Request request `json:"request"`
}

type collection struct {
	Info collectionInfo `json:"info"`
	Item []item         `json:"item"`
}

// Generator builds a Postman Collection v2.1 document from a Pucora ServiceConfig.
type Generator struct {
	// Host is the base host used in generated request URLs. Defaults to "localhost".
	Host string
	// Port is the port used in generated request URLs. Defaults to "8080".
	Port string
}

// Generate walks cfg.Endpoints and produces a Postman Collection v2.1 JSON document.
func (g Generator) Generate(cfg *config.ServiceConfig) ([]byte, error) {
	host := g.Host
	if host == "" {
		host = "localhost"
	}
	port := g.Port
	if port == "" {
		port = "8080"
	}

	col := collection{
		Info: collectionInfo{
			Name:   "Pucora API",
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: make([]item, 0, len(cfg.Endpoints)),
	}

	for _, ep := range cfg.Endpoints {
		method := ep.Method
		if method == "" {
			method = "GET"
		}

		name := fmt.Sprintf("%s %s", method, ep.Endpoint)

		// Extract optional metadata from extra_config.
		if raw, ok := ep.ExtraConfig[Namespace]; ok {
			epCfg, err := parseEndpointConfig(raw)
			if err == nil && epCfg.Name != "" {
				name = epCfg.Name
			}
		}

		// Build path segments (strip leading slash, then split).
		pathStr := strings.TrimPrefix(ep.Endpoint, "/")
		var pathSegments []string
		for _, seg := range strings.Split(pathStr, "/") {
			if seg != "" {
				pathSegments = append(pathSegments, seg)
			}
		}

		rawURL := fmt.Sprintf("http://%s:%s%s", host, port, ep.Endpoint)

		it := item{
			Name: name,
			Request: request{
				Method: method,
				URL: urlObject{
					Raw:  rawURL,
					Host: []string{host},
					Port: port,
					Path: pathSegments,
				},
			},
		}
		col.Item = append(col.Item, it)
	}

	return json.MarshalIndent(col, "", "  ")
}

// parseEndpointConfig converts the raw extra_config value into an EndpointConfig.
func parseEndpointConfig(raw interface{}) (EndpointConfig, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return EndpointConfig{}, err
	}
	var epCfg EndpointConfig
	if err := json.Unmarshal(data, &epCfg); err != nil {
		return EndpointConfig{}, err
	}
	return epCfg, nil
}

// -- CLI command --

var (
	outputFile string

	exportCmd = &cobra.Command{
		Use:     "export-postman",
		Short:   "Exports the Pucora configuration as a Postman Collection v2.1 JSON document.",
		Long:    "Reads the active Pucora configuration file and writes a Postman Collection v2.1 JSON document to stdout or to a file with -o.",
		Example: "pucora export-postman -c pucora.json -o collection.json",
		RunE:    exportFunc,
	}
)

func exportFunc(ccmd *cobra.Command, _ []string) error {
	parser := cmd.GetConfigParser()
	if parser == nil {
		return fmt.Errorf("no config parser available; use ExportCommand as a sub-command of the Pucora root")
	}

	cfg, err := parser.Parse(cmd.GetConfigFlag())
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	g := Generator{}
	data, err := g.Generate(&cfg)
	if err != nil {
		return fmt.Errorf("generating Postman collection: %w", err)
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, data, 0o644)
	}

	_, err = fmt.Fprintln(ccmd.OutOrStdout(), string(data))
	return err
}

// ExportCommand is a cmd.Command that can be added to the Pucora root command.
// It exports the Pucora configuration as a Postman Collection v2.1 JSON document.
var ExportCommand = func() cmd.Command {
	outputFlag := cmd.StringFlagBuilder(&outputFile, "output", "o", "", "Path to the output file (defaults to stdout)")
	cfgFlag := cmd.StringFlagBuilder(new(string), "config", "c", "", "Path to the configuration file")
	return cmd.NewCommand(exportCmd, cfgFlag, outputFlag)
}()
