# Container Example

Test cartage from inside a container. Works with Docker or Podman.

## Setup

From the repo root, build cartage and copy the binary into this directory:

```bash
make build
cp dist/cartage examples/docker/cartage
```

## Run

Start the daemon on the host (in one terminal):

```bash
./dist/cartage serve -v
```

Then start the container (in another terminal):

```bash
cd examples/docker

# With Docker
docker compose build
docker compose run --rm shell

# With Podman
podman build -t cartage-test .
podman run --rm -it -v "$XDG_RUNTIME_DIR/cartage.sock:/run/host/cartage.sock" cartage-test
```

## Test from inside the container

```bash
# Toast notification
notify-send "Hello from container" "It works!"

# Alert dialog (blocks until dismissed on host)
yad --title "Hello" --text "Hello from container" --button ok

# Confirm dialog (exit code 0 = yes, 1 = no)
yad --title "Continue?" --text "Deploy to production?" --button yes --button no
echo $?

# Open a URL on the host
xdg-open https://example.com

# Clipboard
echo "from container" | pbcopy
pbpaste

# Native cartage CLI also works
cartage notify "Test" "Direct CLI"
cartage clipboard copy "hello from container"
cartage clipboard paste
```
