{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
  };

  outputs = { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;

      revSuffix = lib.optionalString (self ? shortRev || self ? dirtyShortRev)
        "-${self.shortRev or self.dirtyShortRev}";

      makePackages = (pkgs: {
        default = pkgs.buildGoModule rec {
          pname = "nebula";
          version = "1.9.5-custom" + revSuffix;

          src = lib.sourceFilesBySuffices ./. [ ".go" ".mod" ".sum" ];

          vendorHash = "sha256-U30rgdAhe69WlwJlGV+inzyu18wEI8RHGjV5DmnWJHQ=";

          subPackages = [ "cmd/nebula" "cmd/nebula-cert" ];

          ldflags = [ "-X main.Build=${version}" ];
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
