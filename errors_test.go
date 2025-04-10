package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	code := "new test"

	// way 1 to make error
	// errors.New()
	e := New(code)
	if e.Code() != code {
		t.Fatalf("want:%s,but:%s", code, e.Code())
		return
	}
	if len(e.Error()) == 0 {
		t.Fatal(e)
	}

	// way 2 to make error
	// errors.As()
	eAs := As(e, "this is value copy for As, not same as before")
	if eAs.Code() != code {
		t.Fatalf("want:%s,but:%s", code, eAs.Code())
		return
	}
	if eAs == e {
		t.Fatalf("%s, not same instance:%s", eAs, e)
		return
	}

	// way 3 to make error
	// errors.ParseError(error)
	eParseErr := ParseError(e)
	if eParseErr.Code() != code {
		t.Fatalf("want:%s,but:%s", code, eParseErr.Code())
		return
	}
	if eParseErr != e {
		t.Fatalf("%s, not same instance:%s", eParseErr, e)
		return
	}

	// way 4 to make error
	// errors.Parse(string)
	eParse := Parse(e.Error())
	if eParse.Code() != code {
		t.Fatalf("want:%s,but:%s", code, eParse.Code())
		return
	}
}

var equalTests = []struct {
	err1 Error
	err2 error
	out  bool
}{
	{New("New"), New("New"), true},
	{New("New"), ParseError(New("New")), true},
	{New("New"), ParseError(errors.New("New")), true},
	{New("New"), As(New("New")), true},
	{New("New"), As(errors.New("New")), true},
	{New("New"), As(New("New"), "reason"), true},
	{New("New"), As(errors.New("New"), "reason"), true},
	{ParseError(New("ParseErr")), ParseError(New("ParseErr")), true},
	{ParseError(New("ParseErr")), ParseError(errors.New("ParseErr")), true},
	{ParseError(New("ParseErr")), As(New("ParseErr")), true},
	{ParseError(New("ParseErr")), As(errors.New("ParseErr")), true},
	{ParseError(New("ParseErr")), As(New("ParseErr"), "reason"), true},
	{ParseError(New("ParseErr")), As(errors.New("ParseErr"), "reason"), true},
}

func TestEqual(t *testing.T) {
	for index, test := range equalTests {
		if test.err1.Equal(test.err2) != test.out {
			t.Fatalf("usercase %d,want:%t,but:%t", index, !test.out, test.out)
			return
		}
		if Equal(test.err1, test.err2) != test.out {
			t.Fatalf("usercase %d,want:%t,but:%t", index, !test.out, test.out)
			return
		}
	}
}

func TestAs(t *testing.T) {
	err1 := New("test")
	err2 := New("test")

	outErr1 := As(err1, "test", "test")
	outErr2 := err2.As("test", "test")
	outErr3 := As(err1, 123, 456)
	outErr4 := err2.As(123, 456)
	if len(outErr1.Error()) == 0 {
		t.Fatal(outErr1)
	}
	if len(outErr2.Error()) == 0 {
		t.Fatal(outErr2)
	}
	if len(outErr3.Error()) == 0 {
		t.Fatal(outErr3)
	}
	if len(outErr4.Error()) == 0 {
		t.Fatal(outErr4)
	}
	fmt.Println("As1:", outErr1)
	fmt.Println("As2:", outErr2)
	fmt.Println("As3:", outErr3)
	fmt.Println("As4:", outErr4.As("call two as"))
	fmt.Println("As5:", outErr4)
}

func TestError(t *testing.T) {
	err1 := New("test")
	err2 := New("test")

	outErr1 := As(err1, "test", "test")
	outErr2 := err2.As("test", "test")
	outErr3 := As(err1, 123, 456)
	outErr4 := err2.As(123, 456)
	if len(outErr1.Error()) == 0 {
		t.Fatal(outErr1)
	}
	if len(outErr2.Error()) == 0 {
		t.Fatal(outErr2)
	}
	if len(outErr3.Error()) == 0 {
		t.Fatal(outErr2)
	}
	if len(outErr4.Error()) == 0 {
		t.Fatal(outErr4)
	}
	// TODO: concurrency testing
	if err1 == outErr1 {
		t.Fatal("need return another error")
	}
	if err2 == outErr2 {
		t.Fatal("need return another error")
	}
	jsonIndent, err := json.MarshalIndent(err1.(*errImpl), "", "	")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Err:", outErr4.Error())
	fmt.Println("Err:", err1.(*errImpl))
	fmt.Println("Err:", err1.As(err2))
	fmt.Println("Err:", As(New("two caller without args")))
	fmt.Println("Json Indent:", string(jsonIndent))
	fmt.Printf("Stack output:%+v\n", err1.Stack())
}
