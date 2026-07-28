// Package ir defines the language-agnostic intermediate representation of a
// generated CLI.
package ir

// A Model is the IR of one input .proto file.
type Model struct {
	// PluginVersion is the version of protoc-gen-cli that produced the model.
	PluginVersion string `json:"plugin_version,omitempty"`
	ProtoFile     string `json:"proto_file,omitempty"`
	// GeneratedFilenamePrefix is the file's output path, minus extension.
	GeneratedFilenamePrefix string      `json:"generated_filename_prefix,omitempty"`
	ProtoPackage            string      `json:"proto_package,omitempty"`
	FileOptions             FileOptions `json:"file_options,omitzero"`
	Services                []*Service  `json:"services,omitempty"`
}

// FileOptions carries the input file's per-language options, e.g. go_package.
type FileOptions struct {
	// GoPackageName is the file's go_package package identifier.
	GoPackageName string `json:"go_package_name,omitempty"`
	// GoImportPath is the file's go_package import path.
	GoImportPath string `json:"go_import_path,omitempty"`
	// GoDescriptorName is protogen's File_<path>_proto identifier for the file.
	GoDescriptorName string `json:"go_descriptor_name,omitempty"`
}

// A Service is one service block: a command with one subcommand per RPC.
type Service struct {
	ProtoName string     `json:"proto_name,omitempty"`
	GoName    string     `json:"go_name,omitempty"`
	Name      string     `json:"name,omitempty"`
	ShortHelp string     `json:"short_help,omitempty"`
	LongHelp  string     `json:"long_help,omitempty"`
	Commands  []*Command `json:"commands,omitempty"`
}

// A Command is one RPC method rendered as a subcommand.
type Command struct {
	ProtoName string   `json:"proto_name,omitempty"`
	GoName    string   `json:"go_name,omitempty"`
	Name      string   `json:"name,omitempty"`
	ShortHelp string   `json:"short_help,omitempty"`
	LongHelp  string   `json:"long_help,omitempty"`
	Input     *Request `json:"input,omitempty"`
	// Output is the response message's full proto name.
	Output          string `json:"output,omitempty"`
	ClientStreaming bool   `json:"client_streaming,omitempty"`
	ServerStreaming bool   `json:"server_streaming,omitempty"`
	// Params is ordered: a message field's own param immediately precedes
	// the params derived from its fields..
	Params []*Param `json:"params,omitempty"`
	View   *View    `json:"view,omitempty"`
}

// A Request identifies an RPC's request message.
type Request struct {
	FullName      string `json:"full_name,omitempty"`
	ProtoFile     string `json:"proto_file,omitempty"`
	ProtoPackage  string `json:"proto_package,omitempty"`
	GoName        string `json:"go_name,omitempty"`
	GoImportPath  string `json:"go_import_path,omitempty"`
	GoPackageName string `json:"go_package_name,omitempty"`
}

// A Param is one command input derived from a request field.
type Param struct {
	ProtoPath string `json:"proto_path,omitempty"`
	Name      string `json:"name,omitempty"`
	Bind      Bind   `json:"bind,omitempty"`
	// Repeated marks a list field's param: each use appends one element.
	Repeated bool `json:"repeated,omitempty"`
	// Map marks a map field's param: each use adds one key=value entry.
	Map bool `json:"map,omitempty"`
	// EnumValues lists an enum field's value names.
	EnumValues []string `json:"enum_values,omitempty"`
	// Oneof names the field's containing oneof; empty for synthetic proto3-optional oneofs.
	Oneof string `json:"oneof,omitempty"`
}

// A Bind represents the JSON type a param's argument parses into.
type Bind string

const (
	BindString Bind = "string"
	BindBool   Bind = "bool"
	BindInt    Bind = "int"
	BindUint   Bind = "uint"
	BindFloat  Bind = "float"
	BindJSON   Bind = "json" // a message as one JSON document
)

// A View projects a response message for display.
type View struct {
	FullName string       `json:"full_name,omitempty"`
	Fields   []*ViewField `json:"fields,omitempty"`
	Lists    []*ViewList  `json:"lists,omitempty"`
}

// A ViewList projects one repeated message field.
type ViewList struct {
	// FullName is the element message's full proto name.
	FullName string       `json:"full_name,omitempty"`
	Label    string       `json:"label,omitempty"`
	Path     string       `json:"path,omitempty"`
	Fields   []*ViewField `json:"fields,omitempty"`
}

// A ViewField is one field of a view: Label derives from the proto field
// path, Path addresses the message's protojson output in gjson syntax.
type ViewField struct {
	Label string `json:"label,omitempty"`
	Path  string `json:"path,omitempty"`
}
