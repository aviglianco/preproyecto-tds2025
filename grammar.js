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
    // ────────────────────────────────────────────────────────────────────────────
    // Entry points
    // ────────────────────────────────────────────────────────────────────────────
    source_file: ($) =>
      choice(seq(choice($._void_type, $._int_type, $._bool_type), $.main)),

    main: ($) => seq("main", field("args", $.args), field("block", $.block)),

    // ────────────────────────────────────────────────────────────────────────────
    // Function arguments
    // ────────────────────────────────────────────────────────────────────────────
    arg: (_$) => "arg",
    args: ($) => seq("(", optional(seq(repeat(seq($.arg, ",")), $.arg)), ")"),

    // ────────────────────────────────────────────────────────────────────────────
    // Types
    // ────────────────────────────────────────────────────────────────────────────
    _void_type: (_$) => "void",
    _bool_type: (_$) => "bool",
    _int_type: (_$) => "int",
    _type: ($) => choice($._int_type, $._bool_type),

    // ────────────────────────────────────────────────────────────────────────────
    // Blocks & statements
    // ────────────────────────────────────────────────────────────────────────────
    block: ($) =>
      seq("{", repeat(seq(field("statement", $._statement), ";")), "}"),

    _statement: ($) =>
      choice(
        $.return_statement,
        "skip",
        $.declaration_statement,
        $.assignment_statement,
        $._expression
      ),

    declaration_statement: ($) =>
      seq(field("type", $._type), field("identifier", $.identifier)),

    assignment_statement: ($) =>
      seq(
        field("identifier", $.identifier),
        "=",
        field("value", $._expression)
      ),

    return_statement: ($) =>
      seq("return", optional(field("value", $._expression))),

    // ────────────────────────────────────────────────────────────────────────────
    // Expressions
    // ────────────────────────────────────────────────────────────────────────────
    _expression: ($) => choice($._exp, seq("(", $._expression, ")")),

    _exp: ($) => choice($._int_operation, $.num, $._bool_const, $.identifier),

    _int_operation: ($) => choice($.int_proc, $.int_div, $.int_sum, $.int_sub),

    int_proc: ($) =>
      prec.right(
        1,
        seq(field("left", $._expression), "*", field("right", $._expression))
      ),
    int_div: ($) =>
      prec.right(
        2,
        seq(field("left", $._expression), "/", field("right", $._expression))
      ),
    int_sum: ($) =>
      prec.right(
        3,
        seq(field("left", $._expression), "+", field("right", $._expression))
      ),
    int_sub: ($) =>
      prec.right(
        4,
        seq(field("left", $._expression), "-", field("right", $._expression))
      ),

    // ────────────────────────────────────────────────────────────────────────────
    // Terminals
    // ────────────────────────────────────────────────────────────────────────────
    identifier: (_$) => /[a-z][a-z,0-9]*/,

    true: (_$) => "true",
    false: (_$) => "false",
    _bool_const: ($) => choice($.true, $.false),

    num: (_$) => /\d+/,
  },
});
