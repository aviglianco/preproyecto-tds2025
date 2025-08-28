/**
 * @file Parser grammar for tree-sitter
 * @author Agus
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

module.exports = grammar({
  name: "parser",

  rules: {
    source_file: $ => choice(seq("void", $.main), "pepe"),
    main: $ => seq("main", $.args, $._block),
    args: _$ => seq("(", repeat(seq("arg", ",")), ")"),
    _block: $ => seq("{", repeat(seq($._statement, ";")), "}"),
    _statement: $ => (choice($.return_statement, "skip", $._intexp1)),

    // TODO: remove parenthesis by using precedence and asociativity
    _int_operation: $ => choice(
      $.int_proc,
      $.int_div,
      $.int_sum,
      $.int_sub,
    ),

    int_proc: $ => prec.right(1, seq($._intexp1, "*", $._intexp1)),
    int_div: $ => prec.right(2, seq($._intexp1, "/", $._intexp1)),
    int_sum: $ => prec.right(3, seq($._intexp1, "+", $._intexp1)),
    int_sub: $ => prec.right(4, seq($._intexp1, "-", $._intexp1)),

    _intexp2: $ => choice($._int_operation, $.num),
    _intexp1: $ => choice($._intexp2, seq("(", $._intexp1, ")")),
    num: _$ => /\d+/,
    return_statement: $ => seq("return", $._intexp1)
  }
});
