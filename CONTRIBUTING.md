# Contributing to `gquil`

Thank you for your interest in contributing to `gquil`!

This tool was borne of my frustration with the existing tooling for working with GraphQL schemas at the command line, so is currently shaped by my own personal preferences and tastes, but I want it to be generally useful for people working with GraphQL on a daily basis.

## Code of conduct

This project has a [code of conduct](./CODE_OF_CONDUCT.md), please observe it.

## Providing feedback

I would love to hear from you about:

- Bugs you encounter while using `gquil`
- Things you were confused by in the behavior of the tool or documentation
- Things you wish the tool would do that it doesn't

You can report any of these kinds of feedback via a GitHub [issue](https://github.com/benweint/gquil/issues).

### What to include

When reporting an issue, please include:

- The version of `gquil` you're using (`gquil version`)
- The exact invocation of `gquil` you tried (or wanted) to run
- Any input schemas necessary to reproduce the problem you're reporting

## Development

### Clone it

```
git clone git@github.com:benweint/gquil.git
```

### Install tools

`gquil` is implemented in Go, so you'll need to have a version of Go installed to build or contribute to it. It uses [`golangci-lint`](https://github.com/golangci/golangci-lint) for linting.

I use [mise](https://mise.jdx.dev/) for managing my local Go & golangci-lint versions. If you do too:

```
cd gquil
mise install

# Check that the versions match what's in `.mise.toml`
go version
golangci-lint --version
```

Otherwise, check `.mise.toml` for the current Go and golangci-lint versions to use.

### Running the tests

```
go test ./...
```

### Running the linter

```
golangci-lint run ./...
```

### Running tests & lints together

```
make check
```

