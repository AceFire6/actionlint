package actionlint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseExpressionSyntaxOK(t *testing.T) {
	testCases := []struct {
		what     string
		input    string
		expected ExprNode
	}{
		// simple expressions
		{
			what:     "null literal",
			input:    "null",
			expected: &NullNode{},
		},
		{
			what:     "boolean literal true",
			input:    "true",
			expected: &BoolNode{Value: true},
		},
		{
			what:     "boolean literal false",
			input:    "false",
			expected: &BoolNode{Value: false},
		},
		{
			what:     "integer literal",
			input:    "711",
			expected: &IntNode{Value: 711},
		},
		{
			what:     "negative integer literal",
			input:    "-10",
			expected: &IntNode{Value: -10},
		},
		{
			what:     "zero integer literal",
			input:    "0",
			expected: &IntNode{Value: 0},
		},
		{
			what:     "hex integer literal",
			input:    "0x1f",
			expected: &IntNode{Value: 0x1f},
		},
		{
			what:     "negative hex integer literal",
			input:    "-0xaf",
			expected: &IntNode{Value: -0xaf},
		},
		{
			what:     "hex integer zero",
			input:    "0x0",
			expected: &IntNode{Value: 0x0},
		},
		{
			what:     "float literal",
			input:    "1234.567",
			expected: &FloatNode{Value: 1234.567},
		},
		{
			what:     "float literal smaller than 1",
			input:    "0.567",
			expected: &FloatNode{Value: 0.567},
		},
		{
			what:     "float literal zero",
			input:    "0.0",
			expected: &FloatNode{Value: 0.0},
		},
		{
			what:     "negative float literal",
			input:    "-1234.567",
			expected: &FloatNode{Value: -1234.567},
		},
		{
			what:     "float literal with exponent part",
			input:    "12e3",
			expected: &FloatNode{Value: 12e3},
		},
		{
			what:     "float literal with negative exponent part",
			input:    "-99e-1",
			expected: &FloatNode{Value: -99e-1},
		},
		{
			what:     "float literal with fraction and exponent part",
			input:    "1.2e3",
			expected: &FloatNode{Value: 1.2e3},
		},
		{
			what:     "float literal with fraction and negative exponent part",
			input:    "-0.123e-12",
			expected: &FloatNode{Value: -0.123e-12},
		},
		{
			what:     "float zero value with exponent part",
			input:    "0e3",
			expected: &FloatNode{Value: 0e3},
		},
		{
			what:     "string literal",
			input:    "'hello, world'",
			expected: &StringNode{Value: "hello, world"},
		},
		{
			what:     "empty string literal",
			input:    "''",
			expected: &StringNode{Value: ""},
		},
		{
			what:     "string literal with escapes",
			input:    "'''hello''world'''",
			expected: &StringNode{Value: "'hello'world'"},
		},
		{
			what:     "string literal with non-ascii chars",
			input:    "'こんにちは＼(^o^)／世界😊'",
			expected: &StringNode{Value: "こんにちは＼(^o^)／世界😊"},
		},
		{
			what:     "variable",
			input:    "github",
			expected: &VariableNode{Name: "github"},
		},
		{
			what:  "< operator",
			input: "0 < 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindLess,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "<= operator",
			input: "0 <= 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindLessEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "> operator",
			input: "0 > 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindGreater,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  ">= operator",
			input: "0 >= 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindGreaterEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "== operator",
			input: "0 == 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "!= operator",
			input: "0 != 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindNotEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "&& operator",
			input: "true && false",
			expected: &LogicalOpNode{
				Kind:  LogicalOpNodeKindAnd,
				Left:  &BoolNode{Value: true},
				Right: &BoolNode{Value: false},
			},
		},
		{
			what:  "|| operator",
			input: "true || false",
			expected: &LogicalOpNode{
				Kind:  LogicalOpNodeKindOr,
				Left:  &BoolNode{Value: true},
				Right: &BoolNode{Value: false},
			},
		},
		{
			what:     "nested value",
			input:    "(42)",
			expected: &IntNode{Value: 42},
		},
		{
			what:     "very nested value",
			input:    "((((((((((((((((((42))))))))))))))))))",
			expected: &IntNode{Value: 42},
		},
		{
			what:  "object property dereference",
			input: "a.b",
			expected: &ObjectDerefNode{
				Receiver: &VariableNode{Name: "a"},
				Property: "b",
			},
		},
		{
			what:  "nested object property dereference",
			input: "a.b.c.d",
			expected: &ObjectDerefNode{
				Property: "d",
				Receiver: &ObjectDerefNode{
					Property: "c",
					Receiver: &ObjectDerefNode{
						Property: "b",
						Receiver: &VariableNode{Name: "a"},
					},
				},
			},
		},
		{
			what:  "array element dereference",
			input: "a.*",
			expected: &ArrayDerefNode{
				Receiver: &VariableNode{Name: "a"},
			},
		},
		{
			what:  "nested array element dereference",
			input: "a.*.*.*",
			expected: &ArrayDerefNode{
				Receiver: &ArrayDerefNode{
					Receiver: &ArrayDerefNode{
						Receiver: &VariableNode{Name: "a"},
					},
				},
			},
		},
	}

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(VariableNode{}),
		cmpopts.IgnoreUnexported(NullNode{}),
		cmpopts.IgnoreUnexported(BoolNode{}),
		cmpopts.IgnoreUnexported(IntNode{}),
		cmpopts.IgnoreUnexported(FloatNode{}),
		cmpopts.IgnoreUnexported(StringNode{}),
		cmpopts.IgnoreUnexported(ObjectDerefNode{}),
		cmpopts.IgnoreUnexported(ArrayDerefNode{}),
		cmpopts.IgnoreUnexported(IndexAccessNode{}),
		cmpopts.IgnoreUnexported(NotOpNode{}),
		cmpopts.IgnoreUnexported(CompareOpNode{}),
		cmpopts.IgnoreUnexported(LogicalOpNode{}),
		cmpopts.IgnoreUnexported(FuncCallNode{}),
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			l := NewExprLexer()
			tok, _, err := l.Lex(tc.input + "}}")
			if err != nil {
				t.Fatal(err)
			}

			p := NewExprParser()
			n, err := p.Parse(tok)
			if err != nil {
				t.Fatal("Parse error:", err)
			}

			if !cmp.Equal(tc.expected, n, opts...) {
				t.Fatalf("wanted:\n%#v\n\nbut got:\n%#v\n\ndiff:\n%s\n", tc.expected, n, cmp.Diff(tc.expected, n, opts...))
			}
		})
	}
}
