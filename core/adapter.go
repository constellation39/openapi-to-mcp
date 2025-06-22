package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/constellation39/openapi-to-mcp/core/session"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/pb33f/libopenapi"
	v3base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

var nameReplacer = strings.NewReplacer(
	" ", "_", "-", "_", "/", "_", ".", "_",
	"{", "", "}", "", ":", "_", "?", "", "&", "and",
	"=", "_eq_", "%", "_pct_",
)

func sanitizeToolName(s string) string {
	s = nameReplacer.Replace(strings.ToLower(s))
	s = strings.Map(func(r rune) rune {
		if r == '_' {
			// skip multiple '_'
			if len(s) > 0 && s[len(s)-1] == '_' {
				return -1
			}
		}
		return r
	}, s)

	s = strings.Trim(s, "_")
	if s == "" {
		return "unnamed_tool"
	}
	return s
}

var pathVarRe = regexp.MustCompile(`\{([^}]+)}`)

func parsePathTmpl(tmpl string) []string {
	matches := pathVarRe.FindAllStringSubmatch(tmpl, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

func NewToolHandlerFromOp(
	baseURL, pathTmpl, method string,
	paramIn map[string]string,
	hasBody bool,
	extraHeaders map[string]string,
) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	pathVars := parsePathTmpl(pathTmpl)

	return func(ctx context.Context, call mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		cli := http.DefaultClient

		if LoadEnv("USE_COOKIE", "true") == "true" {
			if cs := server.ClientSessionFromContext(ctx); cs != nil {
				if st, ok := session.Instance().GetSession(cs.SessionID()); ok && st.Client != nil {
					cli = st.Client
				}
			}
		}

		pathVals := make(map[string]any, len(pathVars))
		queryVals := neturl.Values{}
		headerVals := http.Header{}
		var bodyVal any

		if raw, ok := call.Params.Arguments.(map[string]any); ok {
			for k, v := range raw {
				switch paramIn[k] {
				case "path":
					pathVals[k] = v
				case "query", "":
					queryVals.Add(k, fmt.Sprintf("%v", v))
				case "header":
					headerVals.Add(k, fmt.Sprintf("%v", v))
				case "cookie":
					headerVals.Add("Cookie", fmt.Sprintf("%s=%v", k, v))
				case "body":
					bodyVal = v
				}
			}
		}

		var sb strings.Builder
		sb.Grow(len(baseURL) + len(pathTmpl) + 32)
		sb.WriteString(baseURL)

		cur := pathTmpl
		for _, v := range pathVars {
			ph := "{" + v + "}"
			idx := strings.Index(cur, ph)
			sb.WriteString(cur[:idx])
			sb.WriteString(fmt.Sprintf("%v", pathVals[v]))
			cur = cur[idx+len(ph):]
		}
		sb.WriteString(cur)

		if qs := queryVals.Encode(); qs != "" {
			if strings.ContainsRune(sb.String(), '?') {
				sb.WriteByte('&')
			} else {
				sb.WriteByte('?')
			}
			sb.WriteString(qs)
		}
		finalURL := sb.String()

		var bodyReader io.Reader
		if hasBody && bodyVal != nil {
			b, err := json.Marshal(bodyVal)
			if err != nil {
				return mcp.NewToolResultText("marshal body: " + err.Error()), nil
			}
			bodyReader = bytes.NewReader(b)
		}

		req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), finalURL, bodyReader)
		if err != nil {
			return mcp.NewToolResultText("new request: " + err.Error()), nil
		}
		if bodyReader != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}
		for k, vs := range headerVals {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}

		/********** 3. 用 “会话级” http.Client 发送请求 **********/
		resp, err := cli.Do(req)
		if err != nil {
			return mcp.NewToolResultText("http do: " + err.Error()), nil
		}
		defer resp.Body.Close()

		rb, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultText(string(rb)), nil
	}
}

// -------------------------------------------------------------------
// 3. 其它辅助函数优化
// -------------------------------------------------------------------

// 简化分支查找
func lookupMethodOperation(item *v3high.PathItem) (string, *v3high.Operation) {
	m := map[string]*v3high.Operation{
		"Get":     item.Get,
		"Put":     item.Put,
		"Post":    item.Post,
		"Delete":  item.Delete,
		"Options": item.Options,
		"Head":    item.Head,
		"Patch":   item.Patch,
		"Trace":   item.Trace,
	}
	for k, v := range m {
		if v != nil {
			return k, v
		}
	}
	return "", nil
}

// mergeParameters 保持顺序并避免不必要的复制
func mergeParameters(pathParams, opParams []*v3high.Parameter) []*v3high.Parameter {
	if len(pathParams) == 0 {
		return opParams
	}
	if len(opParams) == 0 {
		return pathParams
	}
	seen := make(map[string]struct{}, len(pathParams)+len(opParams))
	out := make([]*v3high.Parameter, 0, len(pathParams)+len(opParams))

	for _, p := range pathParams {
		if p != nil {
			seen[p.Name] = struct{}{}
			out = append(out, p)
		}
	}
	for _, p := range opParams {
		if p != nil {
			if _, dup := seen[p.Name]; dup {
				// op-level 覆盖 path-level
				for i, old := range out {
					if old.Name == p.Name {
						out[i] = p
						break
					}
				}
			} else {
				out = append(out, p)
			}
		}
	}
	return out
}

func convertSchemaToMCP(name string, s *v3base.Schema, required bool) mcp.ToolOption {
	if s == nil {
		return nil
	}
	t := firstType(s.Type)

	common := make([]mcp.PropertyOption, 0, 6)
	if required {
		common = append(common, mcp.Required())
	}
	if s.Description != "" {
		common = append(common, mcp.Description(s.Description))
	}

	switch t {
	case "string":
		if len(s.Enum) > 0 {
			common = append(common, mcp.Enum(yamlNodesToStrs(s.Enum)...))
		}
		if s.MinLength != nil {
			common = append(common, mcp.MinLength(int(*s.MinLength)))
		}
		if s.MaxLength != nil {
			common = append(common, mcp.MaxLength(int(*s.MaxLength)))
		}
		return mcp.WithString(name, common...)

	case "number", "integer":
		if s.Minimum != nil {
			common = append(common, mcp.Min(*s.Minimum))
		}
		if s.Maximum != nil {
			common = append(common, mcp.Max(*s.Maximum))
		}
		return mcp.WithNumber(name, common...)

	case "boolean":
		return mcp.WithBoolean(name, common...)

	case "array":
		itemMap := map[string]any{"type": firstTypeFromDynamic(s.Items)}
		if s.MinItems != nil {
			common = append(common, mcp.MinItems(int(*s.MinItems)))
		}
		if s.MaxItems != nil {
			common = append(common, mcp.MaxItems(int(*s.MaxItems)))
		}
		common = append(common, mcp.Items(itemMap))
		return mcp.WithArray(name, common...)

	case "object":
		common = append(common, mcp.Properties(schemaPropsToMap(s.Properties)))
		return mcp.WithObject(name, common...)
	}

	return mcp.WithString(name, common...)
}

func pickBaseURLFromDoc(doc v3high.Document) string {
	if servers := doc.Servers; len(servers) > 0 {
		return servers[0].URL
	}
	return ""
}

func AddToolFromOpenAPI(
	mcpServer *server.MCPServer,
	baseURL string,
	extraHeaders map[string]string,
	v3Model *libopenapi.DocumentModel[v3high.Document]) {
	doc := v3Model.Model
	if baseURL == "" {
		baseURL = pickBaseURLFromDoc(doc)
	}
	for it := doc.Paths.PathItems.First(); it != nil; it = it.Next() {
		path := it.Key()
		item := it.Value()
		method, op := lookupMethodOperation(item)
		if op == nil {
			continue
		}

		tool := buildOneTool(path, method, op, item)

		paramIn, hasBody := collectParamLocation(item, op)

		h := NewToolHandlerFromOp(baseURL, path, method, paramIn, hasBody, extraHeaders)

		mcpServer.AddTool(tool, h)
	}
}

func collectParamLocation(item *v3high.PathItem,
	op *v3high.Operation) (map[string]string, bool) {

	mp := map[string]string{}
	for _, p := range mergeParameters(item.Parameters, op.Parameters) {
		if p == nil {
			continue
		}
		mp[p.Name] = p.In // "path" / "query" / "header" / "cookie"
	}

	hasBody := false
	if op.RequestBody != nil && op.RequestBody.Content != nil {
		if _, ok := op.RequestBody.Content.Get("application/json"); ok {
			hasBody = true
			mp["body"] = "body"
		}
	}
	return mp, hasBody
}

func buildOneTool(path, method string,
	op *v3high.Operation, item *v3high.PathItem) mcp.Tool {

	name := op.OperationId
	if name == "" {
		name = sanitizeToolName(fmt.Sprintf("%s_%s", method, path))
	}
	desc := coalesce(op.Description, op.Summary,
		fmt.Sprintf("%s %s", method, path))

	opts := []mcp.ToolOption{mcp.WithDescription(desc)}

	params := mergeParameters(item.Parameters, op.Parameters)
	for _, p := range params {
		if p == nil {
			continue
		}
		if opt := convertParameter(p); opt != nil {
			opts = append(opts, opt)
		}
	}

	if op.RequestBody != nil && op.RequestBody.Content != nil {
		if mt, ok := op.RequestBody.Content.Get("application/json"); ok {
			bodyOpts := convertSchemaToMCP(
				"body", mt.Schema.Schema(), boolVal(op.RequestBody.Required))
			opts = append(opts, bodyOpts)
		}
	}

	return mcp.NewTool(name, opts...)
}

func convertParameter(p *v3high.Parameter) mcp.ToolOption {
	if p.Schema == nil {
		return nil
	}
	required := boolVal(p.Required)
	return convertSchemaToMCP(p.Name, p.Schema.Schema(), required) // 始终只有 1 个
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func boolVal(b *bool) bool {
	return b != nil && *b
}

func firstType(types []string) string {
	if len(types) > 0 {
		return types[0]
	}
	return ""
}

func firstTypeFromDynamic(dyn *v3base.DynamicValue[*v3base.SchemaProxy, bool]) string {
	if dyn == nil {
		return "string"
	}

	switch dyn.N {
	case 0:
		return firstType(dyn.A.Schema().Type)
	}

	return "string"
}

func yamlNodesToStrs(nodes []*yaml.Node) (res []string) {
	for _, n := range nodes {
		res = append(res, n.Value)
	}
	return
}

func schemaPropsToMap(props *orderedmap.Map[string, *v3base.SchemaProxy]) map[string]any {
	out := map[string]any{}
	if props == nil {
		return out
	}
	for el := props.Oldest(); el != nil; el = el.Next() {
		key := el.Key
		val := el.Value.Schema()
		out[key] = map[string]any{"type": firstType(val.Type)}
	}
	return out
}
