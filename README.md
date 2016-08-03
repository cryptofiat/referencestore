# transfer-info

transfer-info implements a server for storing transfer information by blockchain transaction hash.
The information in the transfer info should only be visible to the sender and the receiver.

## Install

0. Tested with Go 1.6
1. Get levelDB `go get github.com/syndtr/goleveldb/leveldb`
2. `./run-daemon.sh`

## Example

```bash
$> curl -X POST -d "Hello world!" http://localhost:8000/7ab54344ab99f90caae7eaa18588b563b2c495286f90a34db2bf19368601e3d8
```

```bash
$> curl http://localhost:8000/7ab54344ab99f90caae7eaa18588b563b2c495286f90a34db2bf19368601e3d8
Hello world!
```
