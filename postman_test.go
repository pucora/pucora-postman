package postman

import (
	"encoding/json"
	"testing"

	"github.com/pucora/lura/v2/config"
)

func TestGenerator_Generate_SingleEndpoint(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name: "test-service",
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/users",
				Method:   "GET",
			},
		},
	}

	g := Generator{Host: "localhost", Port: "8080"}
	data, err := g.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	var col collection
	if err := json.Unmarshal(data, &col); err != nil {
		t.Fatalf("failed to unmarshal generated collection: %v", err)
	}

	if col.Info.Name != "Pucora API" {
		t.Errorf("expected info.name=%q, got %q", "Pucora API", col.Info.Name)
	}
	if col.Info.Schema != "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" {
		t.Errorf("unexpected schema URL: %q", col.Info.Schema)
	}

	if len(col.Item) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Item))
	}

	it := col.Item[0]
	if it.Name != "GET /users" {
		t.Errorf("expected item name=%q, got %q", "GET /users", it.Name)
	}
	if it.Request.Method != "GET" {
		t.Errorf("expected method=GET, got %q", it.Request.Method)
	}
	if it.Request.URL.Raw != "http://localhost:8080/users" {
		t.Errorf("expected raw URL=%q, got %q", "http://localhost:8080/users", it.Request.URL.Raw)
	}
	if len(it.Request.URL.Host) == 0 || it.Request.URL.Host[0] != "localhost" {
		t.Errorf("expected host=[localhost], got %v", it.Request.URL.Host)
	}
	if it.Request.URL.Port != "8080" {
		t.Errorf("expected port=8080, got %q", it.Request.URL.Port)
	}
	if len(it.Request.URL.Path) == 0 || it.Request.URL.Path[0] != "users" {
		t.Errorf("expected path=[users], got %v", it.Request.URL.Path)
	}
}

func TestGenerator_Generate_CustomName(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name: "test-service",
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/products",
				Method:   "POST",
				ExtraConfig: config.ExtraConfig{
					Namespace: map[string]interface{}{
						"name":        "Create Product",
						"description": "Creates a new product.",
					},
				},
			},
		},
	}

	g := Generator{}
	data, err := g.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	var col collection
	if err := json.Unmarshal(data, &col); err != nil {
		t.Fatalf("failed to unmarshal generated collection: %v", err)
	}

	if len(col.Item) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Item))
	}
	if col.Item[0].Name != "Create Product" {
		t.Errorf("expected item name=%q, got %q", "Create Product", col.Item[0].Name)
	}
}

func TestGenerator_Generate_DefaultMethod(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name: "test-service",
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/health",
				// Method intentionally left empty — should default to GET.
			},
		},
	}

	g := Generator{}
	data, err := g.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	var col collection
	if err := json.Unmarshal(data, &col); err != nil {
		t.Fatalf("failed to unmarshal generated collection: %v", err)
	}

	if len(col.Item) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Item))
	}
	if col.Item[0].Request.Method != "GET" {
		t.Errorf("expected default method=GET, got %q", col.Item[0].Request.Method)
	}
}

func TestGenerator_Generate_EmptyEndpoints(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:      "empty-service",
		Endpoints: []*config.EndpointConfig{},
	}

	g := Generator{}
	data, err := g.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	var col collection
	if err := json.Unmarshal(data, &col); err != nil {
		t.Fatalf("failed to unmarshal generated collection: %v", err)
	}
	if len(col.Item) != 0 {
		t.Errorf("expected empty items, got %d", len(col.Item))
	}
}
