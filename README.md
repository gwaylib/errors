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

* Analyze errors position information of Error()
```text
Output：
["test",["errors_test.go:90#errors.TestAs"],["errors_test.go:95#errors.TestAs",123,456]]

Decode: 
["error code", ["runtime stack of New"], ["runtime stack of As", "args of As"...]]
the first one is error code, the second is New, the others are As's called.
```


### Error handling suggestions
* Prioritize handling errors before handling normal logic, where errors are less likely to be ignored and make the program more robust;
* Unless the internal handling result of the error is specified, the error should be returned to the caller;
* If there is no need to return to the caller, a log should be recorded instead of discarding the error
* Normal logic should be avoided from being placed in 'if' as much as possible for easier indentation reading.
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


