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
    source_file: $ => choice(seq(choice($._void_type, $._int_type), $.main)),
    main: $ => seq("main", $.args, $._block),
    arg: _$ => "arg",
    args: $ => seq("(", optional(seq(repeat(seq($.arg, ",")), $.arg)), ")"),

    _void_type: _$ => "void",
    _bool_type: _$ => "bool",
    _int_type: _$ => "int",
    _type: $ => choice($._int_type, $._bool_type),
    _block: $ => seq("{", repeat(seq($._statement, ";")), "}"),
    identifier: _$ => /[a-z][a-z,0-9]*/,
    _statement: $ => (choice($.return_statement, "skip", $.declaration_statement, $.assignment_statement, $._expression)),

    declaration_statement: $ => seq($._type, $.identifier),
    assignment_statement: $ => seq($.identifier, "=", $._expression),

    return_statement: $ => seq("return", optional($._expression)),

    _int_operation: $ => choice(
      $.int_proc,
      $.int_div,
      $.int_sum,
      $.int_sub,
    ),

    int_proc: $ => prec.right(1, seq($._expression, "*", $._expression)),
    int_div: $ => prec.right(2, seq($._expression, "/", $._expression)),
    int_sum: $ => prec.right(3, seq($._expression, "+", $._expression)),
    int_sub: $ => prec.right(4, seq($._expression, "-", $._expression)),

    _exp: $ => choice($._int_operation, $.num, $._bool_const, $.identifier),
    _expression: $ => choice($._exp, seq("(", $._expression, ")")),

    true: _$ => "true",
    false: _$ => "false",
    _bool_const: $ => choice($.true, $.false),

    num: _$ => /\d+/,
  }
});
