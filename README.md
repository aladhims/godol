## Intro

Godol is a utility tool for downloading a file faster by using the power of Go concurrency.

## Installation

From source:

```
$ go get github.com/aladhims/godol/cmd/godol
```

Or manually download it from the [release page](https://github.com/aladhims/godol/releases).


## Usage

Basic (the URL must be specified)

```
$ godol --url=https://example.com/test.jpg
```

With more workers (default: 10) 

```
$ godol --url https://example.com/test.jpg --worker 20
```

Custom file name

```
$ godol --url https://example.com/test.jpg --name foobar.jpg
```

Custom directory

```
$ godol --url https://example.com/test.jpg --dest ~/your/custom/path
```

Full Custom

```
$ godol --url https://example.com/test.jpg --worker 20 --name foobar.jpg --dest ~/your/custom/path
```

Help:

```
$ godol --help
```

![](https://img.shields.io/badge/license-MIT-blue.svg)