# Flashback

Lightweight recovery image installer that handles:

 - Creation of a recovery partition
 - Disk encryption (optional)
 - Factory-reset from the recovery partition

## Install from Source
If you have a Go development environment set up, Go get it:

  ```bash
  $ go get github.com/CanonicalLtd/flashback
  ```

## Build it
```bash
$ cd flashback
$ ./get-deps.sh
$ go install ./...
```
The application will be installed to `$GOPATH/bin`.

## Run it
- Set up the config file, using ```example.yaml``` as a guide.
- Run it:
  ```bash
  $ sudo flashback --config=/path/to/settings.yaml
  ```
