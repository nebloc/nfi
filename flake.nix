{
  description = "Dev shell with Python 3 and virtualenv (no flake-utils)";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }: let
    system = "x86_64-linux"; # Change to your system if needed
    pkgs = import nixpkgs {
      inherit system;
    };
  in {
    packages.${system}.default = pkgs.buildGoModule {
      name = "nfi";
      src = ./.;
      vendorHash = null;
    };
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = [
        pkgs.go
        pkgs.git
        pkgs.curl
        pkgs.bash
        pkgs.python3
        pkgs.python3Packages.virtualenv
      ];

      shellHook = ''
        echo "🐍 Welcome to the Python dev shell!"

        # Optional: Automatically create venv in ./venv
        if [ ! -d venv ]; then
          echo "🔧 Creating virtual environment in ./venv"
          python3 -m venv venv
        fi
        source ./venv/bin/activate
        echo "✅ Virtualenv activated"
      '';
    };
  };
}
