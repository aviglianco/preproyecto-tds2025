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

    varDecl: $ => seq($.type, $.id, "=", $.expr),

    methodDecl: $ => seq(
      choice($.type, "void"),
      $.id,
      "(",
      optional(seq(repeat(seq($.type, $.id, ",")), seq($.type, $.id))),
      ")",
      choice(
        $.block,
        seq("extern", ";")
      )
    ),

    block: $ => seq("{", repeat($.varDecl), repeat($.statement), "}"),

    type: _$ => choice("integer", "bool"),

    statement: $ => choice(
      seq($.id, $.expr, ";"),
      seq($.methodCall, ";"),
      seq("if", "(", $.expr, ")", "then", $.block, optional(seq("else", $.block))),
      seq("while", $.expr, $.block),
      seq("return", optional($.expr))
    ),

    methodCall: $ => seq($.id, "(", optional($.expr), ")"),

    expr: $ => choice(
      $.id,
      $.methodCall,
      $.literal,
      seq($.expr, $.binOp, $.expr),
      seq("-", $.expr),
      seq("!", $.expr),
      seq("(", $.expr, ")")
    ),

    binOp: $ => choice($.arithOp, $.relOp, $.condOp),

    arithOp: _$ => choice("+", "-", "*", "/", "%"),

    relOp: _$ => choice("<", ">", "=="),

    condOp: _$ => choice("&&", "||"),

    literal: $ => choice($.integerLiteral, $.boolLiteral),

    id: $ => seq($.alpha, repeat(choice($.alpha, $.digit))),

    alpha: _$ => /[a-z, A-Z]/,

    digit: _$ => /[0-9]/,

    integerLiteral: $ => repeat1($.digit),

    boolLiteral: _$ => choice(
      "true",
      "false"
    )

  },
});
