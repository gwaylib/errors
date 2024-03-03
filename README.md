### README

An errors recorder package

This design supplements and implements the errors interface of golang,
Used to provide information such as the cause and file location of the error when it occurs, in order to restore the happened scene.

Due to the large amount of data contained in this erroneous design, it is necessary to pay attention to the frequency of calls to avoid affecting system efficiency.

* Example
```text
package main

import "github.com/gwaylib/errors"

func fn1(a int) error {
    if a == 0 {
        // return a common no data error
        return errors.ErrNoData.As(a)
    }
    // other errors with fault position
    return errors.New("not implements").As(a)
}

func fn2(b int) error {
    // call and set the error position, do nothing if 'fn1' return nil
    return errors.As(fn1(b))
}

func main() {
    err := fn2(2)
    if err != nil {
        // Attention, errors.ErrNoData == err may not necessarily hold, the Equal method should be used
        if !errors.ErrNoData.Equal(err) {
            panic(err)
        }
        fmt.Println(err)
    }
}
```

* Analyze errors position information
```text
Output：
["test",["errors_test.go:90#errors.TestAs"],["errors_test.go:95#errors.TestAs",123,456]]

Decode: 
ErrData[0] -- the input of errors.New()
ErrData[1:] -- position information
ErrData[1][0] -- the position information when first calling 'As'
ErrData[1][1:] -- the args when first calling 'As'
```


### Error handling suggestions

*) Prioritize handling errors before handling normal logic, as errors are less likely to be ignored and make the program more robust;
*) Unless the error handling result is clearly defined, errors should always be returned to the caller;
*) If it is not possible to return to the caller, user prompts or logs should be provided instead of discarding errors to fully understand what has happened in the program;
*) Normal logic should not be written in if conditions to ensure good text indentation and reading of the code;
```text
// Suggest
rows, err := db.Query(...)
if err != nil{
    return errors.As(err)
}
defer rows.Close()
...

// Unsuggest
rows, err := db.Query(...)
if err == nil{
    defer rows.Close()
    // ...
} else {
    // handle error or not
}
```

*) Define errors within a small scope, otherwise there is a risk of diffusion
``` text
func Get() (err error){
    // Suggest
    rows, err := db.Query(...)
    if err != nil{
        return errors.As(err)
    }
    defer rows.Close()
  
    if err := rows.Scan(...); err != nil{
        // ...
    }
    ...
}
```


