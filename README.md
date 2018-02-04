# lnchk
Checks all links in a website and returns a summary

## Installation

```
$ go get -u github.com/mrls/lnchk
```

## Usage

```
$ lnchk https://example.com
```

Pipe the results on a tool like `jq` if you want to pretty print the results in the command line

```
$ lnchk https://example.com | jq
```

## Contributing

Bug reports and pull requests are welcome via http://github.com/mrls/lnchk

## License

lnchk is released under the MIT License
