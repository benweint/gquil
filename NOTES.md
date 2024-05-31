## Testing

- [x] high level test coverage
- [ ] test coverage for introspection commands (requires a running server)

## Documentation

- [x] Make `gquil help <subcommand>` show help instead of an error
- [x] Make help text link back to a path for feedback, issues
- [x] Add a CONTRIBUTING.md file with info about development, processes
- [ ] Add examples to the in-tool documentation
- [ ] Add a manpage

## Output / formatting tweaks

- [x] Make `--json` flag format emitted JSON
- [x] Add a `--version` flag

## Argument handling

- [x] Make it possible to read header values from a file ala `curl -H @filename`

## Error handling

- [x] Ensure that schema parse / validation errors actually point back to the source location

## Features

- [x] Add a --named arg to ls fields command
- [ ] Make --interfaces-as-unions work everywhere that --from does
- [ ] Add a --depth-reverse flag to graph filtering options
- [ ] Add a --with-directive flag to filter types, fields by directives
