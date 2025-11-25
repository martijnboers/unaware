{
  pkgs,
  ...
}:
{
  # https://devenv.sh/languages/
  languages.go = {
    enable = true;
  };

  enterShell = ''
    go version
  '';

  tasks."build:linux".exec = "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/unaware-linux-amd64";
  tasks."build:windows".exec = "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/unaware-windows-amd64.exe";
  tasks."build:macos".exec = "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/unaware-darwin-amd64";

}

