blocksafari
===========

blocksafari is a web-based frontend to the blockchain in [btcd](https://github.com/PointCoin/btcd).

**blocksafari is proof-of-concept code from our early work with
  btcd.**

**This code is not suitable for production use and requires major
  refactoring and rewriting to bring it up to speed with current
  proper usage of PointCoin packages and conventions.**

## Installation

#### Build from source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Run the following command to obtain blocksafari, all dependencies, and install it:
  ```$ go get github.com/PointCoin/blocksafari```

- Enter the source directory
  ```cd $GOPATH/src/github.com/PointCoin/blocksafari```

- Copy sample-blocksafari.conf to blocksafari.conf and edit the options.

- Start blocksafari:
  ```blocksafari -C blocksafari.conf```

## Updating

#### Build from Source

- Run the following command to update blocksafari, all dependencies, and install it:
  ```$ go get -u -v github.com/PointCoin/blocksafari/...```

## License

blocksafari is licensed under the liberal ISC License.
