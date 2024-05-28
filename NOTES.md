## Testing

- [x] high level test coverage
- [ ] test coverage for introspection commands (requires a running server)

## Documentation

- [x] Make `gquil help <subcommand>` show help instead of an error
- [ ] Make help text link back to a path for feedback, issues
- [ ] Add examples to the in-tool documentation
- [ ] Add a manpage

## Output / formatting tweaks

- [x] Make `--json` flag format emitted JSON
- [ ] Add a `--version` flag

## Argument handling

- [ ] Make it possible to read header values from a file ala `curl -H @filename`

## Error handling

- [ ] Ensure that schema parse / validation errors actually point back to the source location

## Features

- [ ] Add a --named arg to ls fields command
- [ ] Make --interfaces-as-unions work everywhere that --from does
