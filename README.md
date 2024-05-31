# wazero-helpers: helpful utilities for wazero users

[wazero][1] is a WebAssembly runtime for Go programs, providing a powerful base
for executing Wasm modules. It is relatively low-level though, and there are
certain patterns we see that can be helpful across certain categories of Wasm
modules.

wazero-helpers provides higher level utilities to support some of these use
cases and hopes to provide a helpful base for wazero users that fall into these
categories, while allowing wazero itself to continue to focus on being the an
excellent WebAssembly runtime for Go.

## Development

We use [goyek][2] for defining build tasks. If you have Go installed, you can
run the tasks.

Before sending a PR, it is a good idea to run checks locally with:

```bash
go run ./build check
```

Apply any autoformatting with:

```bash
go run ./build format
```

For VSCode users, it is recommended to use our workspace settings by using
`File > Open Workspace from File...`. All files should autoformat on-save in the same
way as the above command.

[1]: https://wazero.io
[2]: https://github.com/goyek/goyek
