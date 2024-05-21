`gquil` is a tool for introspecting GraphQL schemas on the command line.

It is designed to help make large GraphQL schemas more easily navigable at the command line, and is intended to be used in conjunction with other CLI tools you already use.

It can output information about large GraphQL schemas in several forms:

- A line-delimited format for lists of fields, types, and directives (suitable for direct inspection, or use with `grep`, `sort`, `awk`, etc.)
- A JSON format (suitable for processing with tools like [`jq`](https://github.com/jqlang/jq))
- GraphViz's [DOT language](https://graphviz.org/doc/info/lang.html) for visualization purposes (suitable for use with `dot`)
- GraphQL SDL (suitable for feeding back into `gquil` itself, or using with other GraphQL-related tools)

## Capabilities

### Listing types, fields, and directives

### Visualizing schemas

### Generating GraphQL SDL from an introspection endpoint

### Merging multiple GraphQL SDL files

