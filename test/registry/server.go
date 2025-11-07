// Package main provides a test HTTP server for the Axon registry.
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	registryDir := "."
	if len(os.Args) > 1 {
		registryDir = os.Args[1]
	}

	// Serve static files and web UI
	http.HandleFunc("/", indexHandler(registryDir))
	http.HandleFunc("/api/v1/search", searchHandler(registryDir))
	http.HandleFunc("/api/v1/models/", manifestHandler(registryDir))
	http.HandleFunc("/packages/", packageHandler(registryDir))

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fmt.Printf("🚀 Starting local registry server on http://localhost:%s\n", port)
	fmt.Printf("📁 Registry directory: %s\n", registryDir)
	fmt.Printf("🌐 Web UI: http://localhost:%s\n", port)
	fmt.Printf("🔍 API: http://localhost:%s/api/v1/search?q=<query>\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func indexHandler(registryDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only serve HTML for root path
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// List all models
		models := []map[string]interface{}{}
		manifestsDir := filepath.Join(registryDir, "api/v1/models")

		err := filepath.Walk(manifestsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || !strings.HasSuffix(path, "manifest.yaml") {
				return nil
			}

			// Extract namespace/name/version from path
			relPath, _ := filepath.Rel(manifestsDir, path)
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) >= 3 {
				namespace := parts[0]
				name := parts[1]
				version := strings.TrimSuffix(parts[2], "/manifest.yaml")

				models = append(models, map[string]interface{}{
					"namespace":   namespace,
					"name":        name,
					"version":     version,
					"description": fmt.Sprintf("%s/%s model", namespace, name),
					"manifestUrl": fmt.Sprintf("/api/v1/models/%s/%s/%s/manifest.yaml", namespace, name, version),
				})
			}
			return nil
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("error listing models: %v", err), http.StatusInternalServerError)
			return
		}

		// Render HTML template
		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Axon Local Registry</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 2rem;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        header {
            background: white;
            padding: 2rem;
            border-radius: 12px;
            margin-bottom: 2rem;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        h1 {
            color: #667eea;
            margin-bottom: 0.5rem;
        }
        .subtitle {
            color: #6c757d;
            margin-bottom: 1rem;
        }
        .search-box {
            display: flex;
            gap: 1rem;
            margin-top: 1rem;
        }
        .search-box input {
            flex: 1;
            padding: 0.75rem;
            border: 2px solid #e8ecef;
            border-radius: 8px;
            font-size: 1rem;
        }
        .search-box input:focus {
            outline: none;
            border-color: #667eea;
        }
        .search-box button {
            padding: 0.75rem 2rem;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            cursor: pointer;
            transition: background 0.2s;
        }
        .search-box button:hover {
            background: #5568d3;
        }
        .stats {
            display: flex;
            gap: 2rem;
            margin-top: 1rem;
            color: #6c757d;
        }
        .models-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1.5rem;
        }
        .model-card {
            background: white;
            padding: 1.5rem;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .model-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 8px 12px rgba(0,0,0,0.15);
        }
        .model-header {
            display: flex;
            justify-content: space-between;
            align-items: start;
            margin-bottom: 1rem;
        }
        .model-name {
            font-size: 1.25rem;
            font-weight: 600;
            color: #2c3e50;
        }
        .model-version {
            background: #667eea;
            color: white;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.875rem;
        }
        .model-namespace {
            color: #6c757d;
            font-size: 0.875rem;
            margin-bottom: 0.5rem;
        }
        .model-description {
            color: #6c757d;
            margin-bottom: 1rem;
            line-height: 1.5;
        }
        .model-actions {
            display: flex;
            gap: 0.5rem;
        }
        .btn {
            padding: 0.5rem 1rem;
            border: none;
            border-radius: 6px;
            font-size: 0.875rem;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            transition: background 0.2s;
        }
        .btn-primary {
            background: #667eea;
            color: white;
        }
        .btn-primary:hover {
            background: #5568d3;
        }
        .btn-secondary {
            background: #e8ecef;
            color: #2c3e50;
        }
        .btn-secondary:hover {
            background: #dee2e6;
        }
        .api-info {
            background: white;
            padding: 1.5rem;
            border-radius: 12px;
            margin-top: 2rem;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .api-info h2 {
            color: #667eea;
            margin-bottom: 1rem;
        }
        .api-endpoint {
            background: #f8f9fa;
            padding: 0.75rem;
            border-radius: 6px;
            font-family: 'Courier New', monospace;
            margin: 0.5rem 0;
            word-break: break-all;
        }
        .empty-state {
            text-align: center;
            padding: 3rem;
            background: white;
            border-radius: 12px;
            color: #6c757d;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🧠 Axon Local Registry</h1>
            <p class="subtitle">The Neural Pathway for ML Models</p>
            <div class="search-box">
                <input type="text" id="searchInput" placeholder="Search models (e.g., 'bert', 'vision', 'gpt')..." onkeypress="handleKeyPress(event)">
                <button onclick="searchModels()">Search</button>
            </div>
            <div class="stats">
                <span>📦 <strong>{{len .Models}}</strong> models available</span>
                <span>🔗 <strong>Local Registry</strong></span>
            </div>
        </header>

        <div id="modelsContainer">
            {{if .Models}}
            <div class="models-grid">
                {{range .Models}}
                <div class="model-card">
                    <div class="model-header">
                        <div>
                            <div class="model-namespace">{{.namespace}}</div>
                            <div class="model-name">{{.name}}</div>
                        </div>
                        <span class="model-version">{{.version}}</span>
                    </div>
                    <div class="model-description">{{.description}}</div>
                    <div class="model-actions">
                        <a href="{{.manifestUrl}}" class="btn btn-secondary" target="_blank">View Manifest</a>
                        <button class="btn btn-primary" onclick='installModel("{{.namespace}}/{{.name}}@{{.version}}")'>Install</button>
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="empty-state">
                <h2>No models found</h2>
                <p>Try searching for models or check the registry directory.</p>
            </div>
            {{end}}
        </div>

        <div class="api-info">
            <h2>🔌 API Endpoints</h2>
            <div class="api-endpoint">GET /api/v1/search?q=&lt;query&gt;</div>
            <div class="api-endpoint">GET /api/v1/models/&lt;namespace&gt;/&lt;name&gt;/&lt;version&gt;/manifest.yaml</div>
            <div class="api-endpoint">GET /packages/&lt;package-file&gt;.axon</div>
        </div>
    </div>

    <script>
        function handleKeyPress(event) {
            if (event.key === 'Enter') {
                searchModels();
            }
        }

        function searchModels() {
            const query = document.getElementById('searchInput').value;
            if (!query) {
                location.reload();
                return;
            }

            fetch('/api/v1/search?q=' + encodeURIComponent(query))
                .then(function(response) { return response.json(); })
                .then(function(data) {
                    const container = document.getElementById('modelsContainer');
                    if (data.length === 0) {
                        container.innerHTML = '<div class="empty-state"><h2>No models found</h2><p>Try a different search query.</p></div>';
                        return;
                    }

                    let html = '<div class="models-grid">';
                    data.forEach(function(model) {
                        html += '<div class="model-card">' +
                            '<div class="model-header">' +
                            '<div>' +
                            '<div class="model-namespace">' + model.namespace + '</div>' +
                            '<div class="model-name">' + model.name + '</div>' +
                            '</div>' +
                            '<span class="model-version">' + model.version + '</span>' +
                            '</div>' +
                            '<div class="model-description">' + model.description + '</div>' +
                            '<div class="model-actions">' +
                            '<a href="/api/v1/models/' + model.namespace + '/' + model.name + '/' + model.version + '/manifest.yaml" class="btn btn-secondary" target="_blank">View Manifest</a>' +
                            '<button class="btn btn-primary" onclick="installModel(\'' + model.namespace + '/' + model.name + '@' + model.version + '\')">Install</button>' +
                            '</div>' +
                            '</div>';
                    });
                    html += '</div>';
                    container.innerHTML = html;
                })
                .catch(function(error) {
                    console.error('Search error:', error);
                });
        }

        function installModel(modelSpec) {
            alert('To install this model, run:\\n\\naxon install ' + modelSpec + '\\n\\n\\nMake sure you have configured the registry:\\naxon registry set default http://localhost:8080');
        }
    </script>
</body>
</html>`

		t, err := template.New("index").Parse(tmpl)
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing template: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := struct {
			Models []map[string]interface{}
		}{
			Models: models,
		}
		if err := t.Execute(w, data); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func searchHandler(registryDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		// Search for manifests matching the query
		results := []map[string]interface{}{}
		manifestsDir := filepath.Join(registryDir, "api/v1/models")

		err := filepath.Walk(manifestsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || !strings.HasSuffix(path, "manifest.yaml") {
				return nil
			}

			// Extract namespace/name/version from path
			relPath, _ := filepath.Rel(manifestsDir, path)
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) >= 3 {
				namespace := parts[0]
				name := parts[1]
				version := strings.TrimSuffix(parts[2], "/manifest.yaml")

				// Simple search - check if query matches name or namespace
				if strings.Contains(strings.ToLower(name), strings.ToLower(query)) ||
					strings.Contains(strings.ToLower(namespace), strings.ToLower(query)) {
					results = append(results, map[string]interface{}{
						"namespace":   namespace,
						"name":        name,
						"version":     version,
						"description": fmt.Sprintf("%s/%s model", namespace, name),
					})
				}
			}
			return nil
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("error searching: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func manifestHandler(registryDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract path: /api/v1/models/{namespace}/{name}/{version}/manifest.yaml
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/models/")
		manifestPath := filepath.Join(registryDir, "api/v1/models", path)

		// Check if file exists
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			http.Error(w, "manifest not found", http.StatusNotFound)
			return
		}

		// Serve the YAML file
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, manifestPath)
	}
}

func packageHandler(registryDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract path: /packages/{filename}
		filename := strings.TrimPrefix(r.URL.Path, "/packages/")
		packagePath := filepath.Join(registryDir, "packages", filename)

		// Check if file exists
		if _, err := os.Stat(packagePath); os.IsNotExist(err) {
			http.Error(w, "package not found", http.StatusNotFound)
			return
		}

		// Serve the file
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, packagePath)
	}
}
