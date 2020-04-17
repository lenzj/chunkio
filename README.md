# chunkio

**chunkio** is a golang package that provides functionality for transparently
reading a subset of a stream containing a user defined ending byte sequence.
When the byte sequence is reached an EOF is returned. This sub stream can be
accessed or passed to other routines as standard Reader objects.

## Interface

###Variables

```text
var ErrInvalidKey = errors.New("chunkio: invalid key definition")
```

###Types

```text
type Reader struct {
    // Has unexported fields.
}
    Reader implements chunkio functionality wrapped around an io.Reader object

func NewReader(rd io.Reader) *Reader
    NewReader creates a new chunk reader

func (c *Reader) GetErr() error
    GetErr returns the error status for the current active chunkio stream

func (c *Reader) GetKey() []byte
    GetKey returns the key for the current active chunkio stream

func (c *Reader) Read(p []byte) (int, error)
    Read implements the standard Reader interface allowing chunkio to be used
    anywhere a standard Reader can be used. Read reads data into p. It returns
    the number of bytes read into p. The bytes are taken from at most one Read
    on the underlying Reader, hence n may be less than len(p). When the key is
    reached (EOF for the stream chunk), the count will be zero and err will be
    io.EOF. If the key has been set to nil, the Read function performs exactly
    like the underlying stream Read function (no key scanning).

func (c *Reader) Reset()
    Reset puts the chunkio stream back into a readable state. This can be used
    when the end of a chunk is reached to enable reading the next chunk.

func (c *Reader) SetKey(key []byte) error
    SetKey updates the search key. The search key can also be cleared by
    providing a nil key.
```

## Example usage.

```go
import (
        "bytes"
        "fmt"
        "io/ioutil"
        "git.lenzplace.org/lenzj/chunkio"
        "strings"
)

func ExampleUppercase() {
        example := []byte("the quick {U}brown fox jumps{R} over the lazy dog")
        cio := chunkio.NewReader(bytes.NewReader(example))
        cio.SetKey([]byte("{U}"))
        s1, _ := ioutil.ReadAll(cio)
        cio.Reset()
        cio.SetKey([]byte("{R}"))
        s2, _ := ioutil.ReadAll(cio)
        cio.Reset()
        s3, _ := ioutil.ReadAll(cio)
        fmt.Print(string(s1)+strings.ToUpper(string(s2))+string(s3))
        // Output: the quick BROWN FOX JUMPS over the lazy dog
}
```

## Running the tests

```
$ make check
```

## Contributing

If you have a bugfix, update, issue or feature enhancement the best way to reach
me is by following the instructions in the link below.  Thank you!

<https://blog.lenzplace.org/about/contact.html>


## Versioning

I follow the [SemVer](http://semver.org/) strategy for versioning. The latest
version is listed in the [releases](/lenzj/chunkio/releases) section. 


## License

This project is licensed under a BSD two clause license - see the
[LICENSE](LICENSE) file for details.


<!-- vim:set ts=4 sw=4 et tw=80: -->
