### 使用例子

错误记录器

本设计补充并实现了系统的error接口，

用于发生错误时附带发生错误的原因、位置等信息以便还原现场。

因本错误设计含有比较大的数据量信息，因此需要注意被调用的频率，以避免影响到系统效率。

* 使用例子
```text
package main

import "github.com/gwaylib/errors"

func fn1(a int) error {
   if a == 0 {
       // 返回内置定义的无数据的错误
       return errors.ErrNoData.As(a)
   }
   // 返回其他错误
   //return errors.New("not implements").As(a)
   return errors.New("not implements", a)
}

func fn2(b int) error {
   // 设置一个错误调用位置，若fn1返回nil, 则不设置位置信息
   return errors.As(fn1(b))
}

func main() {
   err := fn2(2)
   if err != nil {
 	  // 注意，errors.ErrNoData == err 会不一定成立, 应使用Equal方法
       if !errors.ErrNoData.Equal(err) {
           panic(err)
       }
       // 处理错误码相等的情况
       fmt.Println(err)
   }
}
```

* Error()定位信息解析
```text
输出：
["test",["errors_test.go:90#errors.TestAs"],["errors_test.go:95#errors.TestAs",123,456]]

解释: 
["错误码", ["New记录的堆栈"], ["As方法记录的堆栈", "As方法记录的参数"...]]
第一个位置是错误天，第二个位置是New的记录，其他位置是As的记录

```

* 读取信息用于自定义格式
```
code := err.Code()   // 读取New()输入的msg值
stack := err.Stack() // 读取各级As()输入的参数值
fmt.Println(code)
for _, s := range stack{ 
    s1 := s.([]interface{})
    fmt.Println(s1...)
}
```


### 错误处理建议

* 优先处理错误，再处理正常逻辑, 此时错误将不容易被忽略而使程序更健壮；
* 除非明确了错误的内部处理结果，否则错误应返回给调用者；
* 若不需向调用者返回, 应当记录日志，而不是丢弃错误
* 正常逻辑尽可能不放在if中，以便于缩进阅读。
```text
// 建议写法
rows, err := db.Query(...)
if err != nil{
    return errors.As(err)
}
defer rows.Close()
...

// 不建议写法
rows, err := db.Query(...)
if err == nil{
    defer rows.Close()
    // ...
} else {
    // 处理错误，或不处理错误
}
```


``` text
// 除非必要，应尽量定义错误在最小范围内，避免定义跨范围使用，它易将错误发散到整个函数而失去可控性
func Get() (err error){
    defer func(){
        if err != nil{
            // 处理错误
        }
    }()

    // ...
}
```


