{ pkgs, ... }: {
  languages.go.enable = true;

  # For tests noexec and builds
  enterShell = ''
    mkdir -p .tmp dist
  '';

  env.TMPDIR = ".tmp";
  env.GOTMPDIR = ".tmp";

  scripts.build-linux.exec = "GOOS=linux GOARCH=amd64 go build -o dist/unaware-linux-amd64 ./cmd";
  scripts.build-windows.exec = "GOOS=windows GOARCH=amd64 go build -o dist/unaware-windows-amd64.exe ./cmd";
  scripts.build-macos.exec = "GOOS=darwin GOARCH=amd64 go build -o dist/unaware-darwin-amd64 ./cmd";
}

