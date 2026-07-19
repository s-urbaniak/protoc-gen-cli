// Package ir defines the language-agnostic intermediate representation of a
// generated CLI.
package ir

// A Model is the IR of one input .proto file.
type Model struct {
	// PluginVersion is the version of protoc-gen-cli that produced the model.
	PluginVersion string      `json:"plugin_version,omitempty"`
	ProtoFile     string      `json:"proto_file,omitempty"`
	ProtoPackage  string      `json:"proto_package,omitempty"`
	FileOptions   FileOptions `json:"file_options,omitzero"`
	Services      []*Service  `json:"services,omitempty"`
}

// FileOptions carries the input file's per-language options, e.g. go_package.
type FileOptions struct {
	// GoPackageName is the file's go_package package identifier.
	GoPackageName string `json:"go_package_name,omitempty"`
	// GoImportPath is the file's go_package import path.
	GoImportPath string `json:"go_import_path,omitempty"`
}

// A Service is one service block: a command group with one command per RPC.
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
	Output string `json:"output,omitempty"`
	// Flags is ordered: a message field's own flag immediately precedes
	// the flags derived from its fields.
	Flags []*Flag `json:"flags,omitempty"`
}

// A Request represents an RPC's request message and the bindings needed to construct it.
type Request struct {
	FullName      string `json:"full_name,omitempty"`
	ProtoFile     string `json:"proto_file,omitempty"`
	ProtoPackage  string `json:"proto_package,omitempty"`
	GoName        string `json:"go_name,omitempty"`
	GoImportPath  string `json:"go_import_path,omitempty"`
	GoPackageName string `json:"go_package_name,omitempty"`
}

// A Flag is a CLI flag derived from a request field.
type Flag struct {
	ProtoPath string `json:"proto_path,omitempty"`
	Name      string `json:"name,omitempty"`
	Bind      Bind   `json:"bind,omitempty"`
	// Repeated marks a list field's flag: each use appends one element.
	Repeated bool `json:"repeated,omitempty"`
	// Map marks a map field's flag: each use adds one key=value entry.
	Map bool `json:"map,omitempty"`
	// EnumValues lists an enum field's value names.
	EnumValues []string `json:"enum_values,omitempty"`
}

// A Bind represents the JSON type a flag's argument parses into.
type Bind string

const (
	BindString Bind = "string" // free text; also bytes, timestamps, durations, enums, field masks
	BindBool   Bind = "bool"
	BindInt    Bind = "int"
	BindUint   Bind = "uint"
	BindFloat  Bind = "float"
	BindJSON   Bind = "json" // a message as one JSON document
)
