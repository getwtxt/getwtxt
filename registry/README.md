# `getwtxt/registry` 

### twtxt Registry Library for Go

`getwtxt/registry` helps you implement twtxt registries in Go.
It uses no third-party dependencies whatsoever, only the standard library,
and has no global state.
Specifying your own `http.Client` for requests is encouraged, with a sensible
default available by passing `nil` to the constructor.

## Using the Library

Just add it to your imports list in the file(s) where it's needed.

```go
import (
  "git.sr.ht/~gbmor/getwtxt/registry"
)
```

## Documentation

The code is commented, so feel free to browse the files themselves. 
Alternatively, the generated documentation can be found at:

[pkg.go.dev/git.sr.ht/~gbmor/getwtxt/registry](https://pkg.go.dev/git.sr.ht/~gbmor/getwtxt/registry)

## Contributions

All contributions are very welcome! Please specify that you are referring to `getwtxt/registry`
when using the following:

* Mailing list (patches, discussion)
  * [https://lists.sr.ht/~gbmor/getwtxt](https://lists.sr.ht/~gbmor/getwtxt)
* Ticket tracker
  * [https://todo.sr.ht/~gbmor/getwtxt](https://todo.sr.ht/~gbmor/getwtxt)

## Notes

* getwtxt - parent project:
  * [sr.ht/~gbmor/getwtxt](https://sr.ht/~gbmor/getwtxt) 

* twtxt repository:
  * [github.com/buckket/twtxt](https://github.com/buckket/twtxt)
* twtxt documentation: 
  * [twtxt.readthedocs.io/en/latest/](https://twtxt.readthedocs.io/en/latest/)
* twtxt registry documentation:
  * [twtxt.readthedocs.io/en/latest/user/registry.html](https://twtxt.readthedocs.io/en/latest/user/registry.html)
