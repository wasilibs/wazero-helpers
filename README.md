# wazero-helpers: helpful utilities for wazero users

[wazero][1] is a WebAssembly runtime for Go programs, providing a powerful base
for executing Wasm modules. It is relatively low-level though, and there are
certain patterns we see that can be helpful across certain categories of Wasm
modules.

wazero-helpers provides higher level utilities to support some of these use
cases and hopes to provide a helpful base for wazero users that fall into these
categories, while allowing wazero itself to continue to focus on being the an
excellent WebAssembly runtime for Go.

[1]: https://wazero.io
