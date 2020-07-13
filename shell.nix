{ pkgsPath ? <nixpkgs> }:

let
  # First we setup our overlays. These are overrides of the official nix packages.
  # We do this to pin the versions we want to use of the software that is in
  # the official nixpkgs repo.
  pkgs = import pkgsPath {
    overlays = [(self: super: {

      go = super.go.overrideAttrs ( old: rec {
        version = "1.14.4";
        src = super.fetchurl {
          url = "https://dl.google.com/go/go${version}.src.tar.gz";
          sha256 = "1105qk2l4kfy1ki9n9gh8j4gfqrfgfwapa1fp38hih9aphxsy4bh";
        };
      });

      go-protobuf = super.go-protobuf.overrideAttrs ( old: rec {
        version = "1.3.5";
        src = super.fetchFromGitHub {
          owner = "golang";
          repo = "protobuf";
          rev = "v${version}";
          sha256 = "1gkd1942vk9n8kfzdwy1iil6wgvlwjq7a3y5jc49ck4lz9rhmgkq";
        };

        modSha256 = "0jjjj9z1dhilhpc8pq4154czrb79z9cm044jvn75kxcjv6v5l2m5";
      });

    })];
  };
in with pkgs; let
  protoc-gen-validate = buildGoModule rec {
    pname = "protoc-gen-validate";
    version = "0.4.0";

    src = fetchFromGitHub {
      owner = "envoyproxy";
      repo = "protoc-gen-validate";
      rev = "v0.4.0";
      sha256 = "0w352i2nlsz069v28q99mz1590c3wba9f55slz51pmgyr9qlil3c";
    };

    modSha256 = "1s5kxj25zw0zwqrdbcq45jv1f8g430n8ijf4c4lax6sismzgwc07";

    subPackages = [ "." ];
  };

in pkgs.mkShell rec {
  name = "horizon";

  # The packages in the `buildInputs` list will be added to the PATH in our shell
  buildInputs = [
    pkgs.go
    pkgs.go-bindata
    pkgs.go-protobuf
    pkgs.protobuf3_11
    pkgs.postgresql_12
    protoc-gen-validate
  ];

  # Extra env vars
  PGHOST = "localhost";
  PGPORT = "5432";
  PGDATABASE = "noop";
  PGUSER = "postgres";
  PGPASSWORD = "postgres";
  DATABASE_URL = "postgresql://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/${PGDATABASE}?sslmode=disable";
}
