package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"
)

type NixFlake struct {
	Description string
	Name        string
	Pkgs        []string
	System      string
	Language    string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <name> <language> [-p packages] [-d description]\n", os.Args[0])
		os.Exit(1)
	}

	name := os.Args[1]
	language := os.Args[2]

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	packages := fs.String("p", "", "Comma-separated list of packages")
	description := fs.String("d", "", "Optional description")

	err := fs.Parse(os.Args[3:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse flags:", err)
		os.Exit(1)
	}

	// Split packages into slice
	var packageList []string
	if *packages != "" {
		packageList = strings.Split(*packages, ",")
	}
	packageList = append(packageList, languagePkgs[language]...)

	if *description == "" {
		*description = "Dev Shell Flake"
	}

	flake := NixFlake{
		Name:        name,
		Description: *description,
		Pkgs:        packageList,
		System:      nixSystem(),
		Language:    language,
	}
	generateTemplates(flake)

}

func generateTemplates(flake NixFlake) {
	tmpl := template.New("Flake.nix")

	tmpl, err := tmpl.Parse(shellHook(flake.Language))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse hook template: %s\n", err)
		os.Exit(1)
	}

	tmpl, err = tmpl.Parse(nix_template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse flake template: %s\n", err)
		os.Exit(1)
	}

	err = tmpl.Execute(os.Stdout, flake)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to render template: %s\n", err)
		os.Exit(1)
	}
}

var languagePkgs = map[string][]string{
	"go":     {"go"},
	"python": {"python3"},
	"rust":   {"rustc", "cargo", "rustfmt", "rust-analyzer", "clippy"},
}

func shellHook(language string) (result string) {
	defer func() {
		result = `{{define "hook"}}` + result + `{{end}}`
	}()

	switch strings.ToLower(language) {
	case "python":
		return python_shell_hook
	case "go":
		return go_shell_hook
	default:
		return ""
	}
}

const python_shell_hook = `
        echo "🐍 Welcome to the {{.Name}} dev shell!"

        # Optional: Automatically create venv in ./venv
        if [ ! -d venv ]; then
          echo "🔧 Creating virtual environment in ./venv"
          python3 -m venv venv
        fi
        source ./venv/bin/activate
        echo "✅ Virtualenv activated"
`

const go_shell_hook = `
        echo "🐹 Welcome to the {{.Name}} dev shell!"

	if [ ! -e go.mod ]; then
        	echo "🔧 Creating go mod file"
		go mod init {{.Name}}
	fi
	`

const nix_template = `
{
  description = "{{.Description}}";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }: let
    system = "{{.System}}"; # Change to your system if needed
    pkgs = import nixpkgs {
      inherit system;
    };
  in {
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [{{range .Pkgs}}
	{{.}}{{end}}
      ];

      shellHook = ''{{template "hook" .}}'';
    };
  };
}
`

func nixSystem() string {
	arch := runtime.GOARCH
	os := runtime.GOOS

	var nixArch string
	switch arch {
	case "amd64":
		nixArch = "x86_64"
	case "arm64":
		nixArch = "aarch64"
	default:
		nixArch = arch // fallback; may need more mappings
	}

	var nixOS string
	switch os {
	case "linux":
		nixOS = "linux"
	case "darwin":
		nixOS = "darwin"
	default:
		nixOS = os // fallback; may need more mappings
	}

	return fmt.Sprintf("%s-%s", nixArch, nixOS)
}
