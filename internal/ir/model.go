// Package ir defines the language-agnostic intermediate representation of a
// generated CLI.
package ir

// A Model is the IR of one input .proto file.
type Model struct {
	// PluginVersion is the version of protoc-gen-cli that produced the model.
	PluginVersion string `json:"plugin_version,omitempty"`
	ProtoFile     string `json:"proto_file,omitempty"`
	// GeneratedFilenamePrefix is the output path without the extension.
	GeneratedFilenamePrefix string      `json:"generated_filename_prefix,omitempty"`
	ProtoPackage            string      `json:"proto_package,omitempty"`
	FileOptions             FileOptions `json:"file_options,omitzero"`
	Services                []*Service  `json:"services,omitempty"`
}

// FileOptions contains the input file's per-language options, for example
// go_package.
type FileOptions struct {
	GoPackageName string `json:"go_package_name,omitempty"`
	GoImportPath  string `json:"go_import_path,omitempty"`
}

// A Service is one service block. It becomes a command with one subcommand
// for each RPC.
type Service struct {
	FullName  string   `json:"full_name,omitempty"`
	ProtoName string   `json:"proto_name,omitempty"`
	Name      string   `json:"name,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	// Deprecated hides the command from help. It still runs.
	Deprecated bool       `json:"deprecated,omitempty"`
	ShortHelp  string     `json:"short_help,omitempty"`
	LongHelp   string     `json:"long_help,omitempty"`
	Commands   []*Command `json:"commands,omitempty"`
}

type Command struct {
	FullName  string   `json:"full_name,omitempty"`
	ProtoName string   `json:"proto_name,omitempty"`
	Name      string   `json:"name,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	// Deprecated hides the subcommand from help. It still runs.
	Deprecated bool        `json:"deprecated,omitempty"`
	ShortHelp  string      `json:"short_help,omitempty"`
	LongHelp   string      `json:"long_help,omitempty"`
	Request    *MessageRef `json:"request,omitempty"`
	Response   string      `json:"response,omitempty"`
	Shape      Shape       `json:"shape,omitempty"`
	// A message field's param precedes its sub-fields' params.
	Params []*Param `json:"params,omitempty"`
	View   *View    `json:"view,omitempty"`
	// ExampleJSON is a sample request in protojson form.
	ExampleJSON string `json:"example_json,omitempty"`
}

type Shape string

const (
	ShapeUnary        Shape = "unary"
	ShapeServerStream Shape = "server_stream"
	ShapeClientStream Shape = "client_stream"
	ShapeBidi         Shape = "bidi"
)

type MessageRef struct {
	FullName      string `json:"full_name,omitempty"`
	ProtoFile     string `json:"proto_file,omitempty"`
	ProtoPackage  string `json:"proto_package,omitempty"`
	GoImportPath  string `json:"go_import_path,omitempty"`
	GoPackageName string `json:"go_package_name,omitempty"`
}

type Param struct {
	ProtoPath string `json:"proto_path,omitempty"`
	Name      string `json:"name,omitempty"`
	Shorthand string `json:"shorthand,omitempty"`
	// Deprecated hides the param from help. It still parses.
	Deprecated bool     `json:"deprecated,omitempty"`
	ShortHelp  string   `json:"short_help,omitempty"`
	Bind       Bind     `json:"bind,omitempty"`
	Repeated   bool     `json:"repeated,omitempty"`
	EnumValues []string `json:"enum_values,omitempty"`
	// Each use adds one key=value entry.
	Map bool `json:"map,omitempty"`
	// Oneof is empty for a synthetic proto3-optional oneof.
	Oneof string `json:"oneof,omitempty"`
}

// A Bind is the JSON type that a param's argument becomes.
// Two binds differ when protojson accepts different text for them.
type Bind string

const (
	BindString Bind = "string"
	BindBool   Bind = "bool"
	BindInt    Bind = "int"
	BindUint   Bind = "uint"
	BindFloat  Bind = "float"
	BindJSON   Bind = "json"
	BindList   Bind = "list"

	BindBytes     Bind = "bytes"      // Base64.
	BindTimestamp Bind = "timestamp"  // RFC 3339, for example 2026-01-30T15:04:05Z.
	BindDuration  Bind = "duration"   // Seconds with an s, for example 3600s.
	BindFieldMask Bind = "field_mask" // Comma-joined field paths.
	BindAny       Bind = "any"        // A JSON document with its own @type.
)

type View struct {
	FullName string       `json:"full_name,omitempty"`
	Fields   []*ViewField `json:"fields,omitempty"`
	Lists    []*ViewList  `json:"lists,omitempty"`
}

type ViewList struct {
	FullName string       `json:"full_name,omitempty"`
	Label    string       `json:"label,omitempty"`
	Path     string       `json:"path,omitempty"`
	Fields   []*ViewField `json:"fields,omitempty"`
}

type ViewField struct {
	Label string `json:"label,omitempty"`
	// Path selects the field from the message's protojson output. RFC 9535 JSONPath.
	Path string `json:"path,omitempty"`
}
