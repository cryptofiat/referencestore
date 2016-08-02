# referencestore
Server for storing transfer reference values by blockchain transaction reference. 
The information in the transfer reference should only be visible to the sender and the receiver.

## Install 

0. Tested with golang 1.6
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
