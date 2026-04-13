{
  description = "Dindoa - Generate ICS calendar files for Dindoa korfbal team matches";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
          version = builtins.readFile ./VERSION;
        in
        {
          dindoa = pkgs.buildGoModule {
            pname = "dindoa";
            version = pkgs.lib.strings.trim version;
            src = ./.;

            # vendorHash is automatically updated by release.sh when Go dependencies change.
            # To manually update: Run `nix build .#dindoa 2>&1 | grep "got:"` and use the hash shown.
            vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

            ldflags = [
              "-s"
              "-w"
            ];

            meta = with pkgs.lib; {
              description = "Generate ICS calendar files for Dindoa korfbal team matches";
              homepage = "https://github.com/torreirow/dindoa";
              license = licenses.mit;
              maintainers = [ ];
            };
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.dindoa);

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gotools
              go-tools
            ];

            shellHook = ''
              echo "Dindoa development environment"
              echo "Go version: $(go version)"
              echo ""
              echo "Available commands:"
              echo "  go build -o dindoa cmd/dindoa/main.go  - Build the binary"
              echo "  go test ./...                          - Run tests"
              echo "  ./dindoa start                         - Start interactive mode"
            '';
          };
        });
    };
}
