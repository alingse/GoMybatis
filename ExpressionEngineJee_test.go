package GoMybatis

import (
	"fmt"
	"testing"
)

func TestExpressionEngineJee_Eval(t *testing.T) {
	//使用go语言判断表达式
	var equalResult bool
	var page = -9
	if page <= 0 && page <= -8 || page > 0 {
		fmt.Println("go=true")
		equalResult = true
	} else {
		fmt.Println("go=false")
		equalResult = false
	}
	//使用ExpressionEngineJee判断表达式
	var m = make(map[string]interface{})
	m["page"] = -9
	var engine = ExpressionEngineJee{}
	var newStr = ".page <= 0 and .page <= -8 or .page > 0"
	result, error := engine.LexerEval(newStr, m, JeeOperation_Marshal_Map)
	if error != nil {
		t.Fatal(error)
	}
	fmt.Println("jeeEngine=", result)
	if equalResult != result {
		t.Fatal("jeeEngine equal != go equal")
	}
}

func TestExpressionEngineJee_LexerAndOrSupport(t *testing.T) {
	var a = 2
	if a > 0 && a > 1 || a < 0 {
		fmt.Println("y")
	} else {
		fmt.Println("n")
	}
	var newStr = ".page <= 0 and .page != 0 or .page <=0"
	var engine = ExpressionEngineJee{}
	var lexerStr = engine.LexerAndOrSupport(newStr)
	fmt.Println(lexerStr)
}

func TestExpressionEngineJee_Eval_null(t *testing.T) {
	var m = make(map[string]interface{})
	m["a"] = 1
	var engine = ExpressionEngineJee{}
	var lexer, _ = engine.Lexer(".a == null")
	var result, error = engine.Eval(lexer, m, JeeOperation_Marshal_Map)
	if error != nil {
		t.Fatal(error)
	}
	fmt.Println(result)
}

func BenchmarkExpressionEngineJee_Eval(b *testing.B) {
	b.StopTimer()
	var m = make(map[string]interface{})
	m["a"] = nil
	var engine = ExpressionEngineJee{}

	//start
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		var lexer, _ = engine.Lexer(".a == null")
		var result, error = engine.Eval(lexer, m, JeeOperation_Marshal_Map)
		if error != nil {
			b.Fatal(error)
		}
		if result == nil {
			b.Fatal("eval fail")
		}
	}
}


