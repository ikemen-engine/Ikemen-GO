{
  description = "Ikemen GO Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixgl.url = "github:nix-community/nixGL";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixgl,
    }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      nixGL = nixgl.packages.${system}.nixGLDefault;
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          go
          pkg-config
          gnumake
          git
          nasm
          yasm
        ];

        buildInputs = with pkgs; [
          SDL2
          libxmp
          ffmpeg
          gtk3
          libGL
          libx11
          vulkan-loader
          nixGL
        ];

        shellHook = ''
          export LD_LIBRARY_PATH="/usr/lib:/usr/lib32:$LD_LIBRARY_PATH"
        '';
      };
    };
}
