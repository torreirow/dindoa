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
            vendorHash = "sha256-KBINEvO1fD6C3LGozXzcDUyfueWvLDTgHig+Lsy46F8=";

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
              # Drop into the user's real login shell with a marked prompt.
              # Nix sets $SHELL to bash, so without this you lose your own
              # configuration, history and completions inside the dev shell.
              #
              # Inert when stdin or stdout is not a tty, and when the guard is
              # already set: that keeps `nix develop -c ...`, scripts and CI
              # untouched, and re-entering does not stack nested shells.
              if [ -t 0 ] && [ -t 1 ] && [ -z "$DINDOA_DEV_SHELL" ]; then
                export DINDOA_DEV_SHELL=1

                _DINDOA_DIR=$(mktemp -d)

                # getent is absent on darwin, hence the two fallbacks.
                _LOGIN_SHELL=$(getent passwd "$USER" 2>/dev/null | cut -d: -f7)
                if [ -z "$_LOGIN_SHELL" ] && [ "$(uname)" = "Darwin" ]; then
                  _LOGIN_SHELL=$(dscl . -read "/Users/$USER" UserShell 2>/dev/null | awk '{print $2}')
                fi
                _LOGIN_SHELL="''${_LOGIN_SHELL:-''${SHELL:-zsh}}"

                case "$_LOGIN_SHELL" in
                  */zsh)
                    # ZDOTDIR trick: source the user's own .zshrc first, then
                    # unset so nested shells behave normally again, and only
                    # then prefix the prompt.
                    cat > "$_DINDOA_DIR/.zshrc" << 'DINDOA_ZSHRC'
[[ -f "$HOME/.zshrc" ]] && source "$HOME/.zshrc"
unset ZDOTDIR
PROMPT="%F{208}(nix-dev)%f $PROMPT"
DINDOA_ZSHRC
                    export ZDOTDIR="$_DINDOA_DIR"
                    ;;
                  */bash)
                    cat > "$_DINDOA_DIR/.bashrc" << 'DINDOA_BASHRC'
[[ -f "$HOME/.bashrc" ]] && source "$HOME/.bashrc"
PS1="\[\e[38;5;208m\](nix-dev)\[\e[0m\] ''${PS1:-\\$ }"
DINDOA_BASHRC
                    _DINDOA_EXEC_ARGS=(--rcfile "$_DINDOA_DIR/.bashrc")
                    ;;
                esac

                exec "$_LOGIN_SHELL" "''${_DINDOA_EXEC_ARGS[@]}"
              fi
            '';
          };
        });
    };
}
