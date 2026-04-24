package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type headerRow struct {
	Name   string   `json:"name" xml:"name"`
	Values []string `json:"values" xml:"values>value"`
}

type responseData struct {
	XMLName    xml.Name    `json:"-" xml:"request"`
	Method     string      `json:"method" xml:"method"`
	Path       string      `json:"path" xml:"path"`
	RawQuery   string      `json:"rawQuery" xml:"rawQuery"`
	Host       string      `json:"host" xml:"host"`
	RemoteAddr string      `json:"remoteAddr" xml:"remoteAddr"`
	Proto      string      `json:"proto" xml:"proto"`
	Headers    []headerRow `json:"headers" xml:"headers>header"`
}

var page = template.Must(template.New("headers").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Request Headers</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 2rem; background: #f7f7f7; color: #222; }
    h1 { margin-bottom: 0.5rem; }
    .card { background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 1rem 1.25rem; margin-bottom: 1rem; }
    table { width: 100%; border-collapse: collapse; background: #fff; }
    th, td { text-align: left; padding: 0.75rem; border-bottom: 1px solid #e5e5e5; vertical-align: top; }
    th { width: 240px; background: #fafafa; }
    code { background: #f1f1f1; padding: 0.1rem 0.3rem; border-radius: 4px; }
    .muted { color: #666; }
  </style>
</head>
<body>
  <h1>Request Inspector</h1>
  <p class="muted">Use this behind your auth proxy to confirm which headers reach the app.</p>

  <div class="card">
    <p><strong>Method:</strong> <code>{{.Method}}</code></p>
    <p><strong>Path:</strong> <code>{{.Path}}</code></p>
    <p><strong>Query:</strong> <code>{{if .RawQuery}}{{.RawQuery}}{{else}}(none){{end}}</code></p>
    <p><strong>Host:</strong> <code>{{.Host}}</code></p>
    <p><strong>Remote Addr:</strong> <code>{{.RemoteAddr}}</code></p>
    <p><strong>Protocol:</strong> <code>{{.Proto}}</code></p>
  </div>

  <table>
    <thead>
      <tr>
        <th>Header</th>
        <th>Value(s)</th>
      </tr>
    </thead>
    <tbody>
      {{range .Headers}}
      <tr>
        <th>{{.Name}}</th>
        <td><code>{{range $index, $value := .Values}}{{if $index}}, {{end}}{{$value}}{{end}}</code></td>
      </tr>
      {{else}}
      <tr>
        <th colspan="2">No headers received</th>
      </tr>
      {{end}}
    </tbody>
  </table>
</body>
</html>`))

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := buildResponseData(r)
		format := negotiateFormat(r)

		switch format {
		case "json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			if err := enc.Encode(data); err != nil {
				http.Error(w, fmt.Sprintf("json error: %v", err), http.StatusInternalServerError)
			}
		case "xml":
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.Write([]byte(xml.Header))
			enc := xml.NewEncoder(w)
			enc.Indent("", "  ")
			if err := enc.Encode(data); err != nil {
				http.Error(w, fmt.Sprintf("xml error: %v", err), http.StatusInternalServerError)
			}
		case "text":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			if _, err := w.Write([]byte(renderText(data))); err != nil {
				http.Error(w, fmt.Sprintf("text error: %v", err), http.StatusInternalServerError)
			}
		default:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := page.Execute(w, data); err != nil {
				http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
			}
		}
	})

	addr := ":" + port
	log.Printf("listening on http://0.0.0.0%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func buildResponseData(r *http.Request) responseData {
	headers := make([]headerRow, 0, len(r.Header))
	names := make([]string, 0, len(r.Header))
	for name := range r.Header {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		headers = append(headers, headerRow{
			Name:   name,
			Values: append([]string(nil), r.Header.Values(name)...),
		})
	}

	return responseData{
		Method:     r.Method,
		Path:       r.URL.Path,
		RawQuery:   r.URL.RawQuery,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
		Proto:      r.Proto,
		Headers:    headers,
	}
}

func negotiateFormat(r *http.Request) string {
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	switch format {
	case "html", "json", "xml", "text", "txt":
		if format == "txt" {
			return "text"
		}
		return format
	}

	accept := strings.ToLower(r.Header.Get("Accept"))
	switch {
	case strings.Contains(accept, "application/json"):
		return "json"
	case strings.Contains(accept, "application/xml"), strings.Contains(accept, "text/xml"):
		return "xml"
	case strings.Contains(accept, "text/plain"):
		return "text"
	default:
		return "html"
	}
}

func renderText(data responseData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Method: %s\n", data.Method)
	fmt.Fprintf(&b, "Path: %s\n", data.Path)
	if data.RawQuery == "" {
		fmt.Fprintf(&b, "Query: (none)\n")
	} else {
		fmt.Fprintf(&b, "Query: %s\n", data.RawQuery)
	}
	fmt.Fprintf(&b, "Host: %s\n", data.Host)
	fmt.Fprintf(&b, "Remote Addr: %s\n", data.RemoteAddr)
	fmt.Fprintf(&b, "Protocol: %s\n", data.Proto)
	b.WriteString("Headers:\n")
	if len(data.Headers) == 0 {
		b.WriteString("  (none)\n")
		return b.String()
	}
	for _, header := range data.Headers {
		fmt.Fprintf(&b, "  %s: %s\n", header.Name, strings.Join(header.Values, ", "))
	}
	return b.String()
}
