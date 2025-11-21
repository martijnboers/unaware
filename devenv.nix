{ ... }:
{
  languages.go.enable = true;

  # For tests noexec and builds
  enterShell = ''
    mkdir -p .tmp dist
  '';

  env.TMPDIR = ".tmp";
  env.GOTMPDIR = ".tmp";

  scripts.build-linux.exec = "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/unaware-linux-amd64";
  scripts.build-windows.exec = "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/unaware-windows-amd64.exe";
  scripts.build-macos.exec = "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/unaware-darwin-amd64";
}
