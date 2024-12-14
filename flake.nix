{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
  };

  outputs = { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;

      revSuffix = lib.optionalString (self ? shortRev || self ? dirtyShortRev)
        "-${self.shortRev or self.dirtyShortRev}";

      makePackages = (pkgs: {
        default = pkgs.buildGoModule rec {
          pname = "nebula";
          version = "1.11.0-custom" + revSuffix;

          src = lib.sourceFilesBySuffices ./. [ ".go" ".mod" ".sum" ];

          vendorHash = "sha256-bTVGIDEjVye1oY/gKKkXgcJYx3AQhF3VcuzdnWSL4NQ=";

          subPackages = [ "cmd/nebula" "cmd/nebula-cert" ];

          ldflags = [ "-X main.Build=${version}" ];

          nativeInstallCheckInputs = with pkgs; [ versionCheckHook ];
          doInstallCheck = true;
          meta.mainProgram = pname;
        };
      }
      );
    in
    builtins.foldl' lib.recursiveUpdate { } (builtins.map
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };

          packages = makePackages pkgs;
        in
        {
          devShells.${system} = packages;
          packages.${system} = packages;
        })
      lib.systems.flakeExposed);
}
