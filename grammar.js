/**
 * @file Parser grammar for tree-sitter
 * @author Agus
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

export default grammar({
  name: "preprojectlang",

  rules: {
    program: $ => seq("program", "{", repeat($.varDecl), repeat($.methodDecl), "}"),

    varDecl: $ => seq($.type, $.id, "=", $.expr, ";"),

    methodDecl: $ => seq(
      choice($.type, "void"),
      $.id,
      "(",
      optional(seq(repeat(seq($.type, $.id, ",")), $.type, $.id)),
      ")",
      choice(
        $.block,
        seq("extern", ";")
      )
    ),

    block: $ => seq("{", repeat($.varDecl), repeat($.statement), "}"),

    type: _$ => choice("integer", "bool"),

    returnStatement: $ => seq("return", optional($.expr), ";"),


    statement: $ => choice(
      seq($.id, "=", $.expr, ";"),
      seq($.methodCall, ";"),
      seq("if", "(", $.expr, ")", "then", $.block, optional(seq("else", $.block))),
      seq("while", $.expr, $.block),
      $.returnStatement,
    ),

    methodCall: $ => seq($.id, "(", optional($.expr), ")"),

    expr: $ => choice(
      $.id,
      $.methodCall,
      $.literal,
      $.binOp,
      prec.right(seq("-", $.expr)),
      prec.right(seq("!", $.expr)),
      seq("(", $.expr, ")")
    ),

    binOp: $ => choice(
      $.sum,
      $.sub,
      $.mul,
      $.div,
      $.rem,
      $.gt,
      $.lt,
      $.eq,
      $.and,
      $.or
    ),
    sum: $ => prec.right(seq($.expr, "+", $.expr)),
    sub: $ => prec.right(seq($.expr, "-", $.expr)),
    mul: $ => prec.right(1, seq($.expr, "*", $.expr)),
    div: $ => prec.right(1, seq($.expr, "/", $.expr)),
    rem: $ => prec.right(1, seq($.expr, "%", $.expr)),
    gt: $ => prec.right(1, seq($.expr, ">", $.expr)),
    lt: $ => prec.right(1, seq($.expr, "<", $.expr)),
    eq: $ => prec.right(1, seq($.expr, "==", $.expr)),
    and: $ => prec.right(1, seq($.expr, "&&", $.expr)),
    or: $ => prec.right(1, seq($.expr, "||", $.expr)),

    arithOp: _$ => choice("+", "-", "*", "/", "%"),

    relOp: _$ => choice("<", ">", "=="),

    condOp: _$ => choice("&&", "||"),

    literal: $ => choice($.integerLiteral, $.boolLiteral),

    id: _$ => /[a-z,A-Z][a-z, A-Z,0-9]*/,

    word: $ => $.id,

    integerLiteral: _$ => /[0-9]+/,

    boolLiteral: _$ => choice(
      "true",
      "false"
    )

  },
});
