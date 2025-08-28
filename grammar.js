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
    _statement: $ => (choice($.return_statement, "skip", $.intexp)),

    // TODO: remove parenthesis by using precedence and asociativity
    int_operation: $ => choice(
      seq("(", $.intexp, "+", $.intexp, ")"),
      seq("(", $.intexp, "/", $.intexp, ")"),
      seq("(", $.intexp, "*", $.intexp, ")"),
      seq("(", $.intexp, "-", $.intexp, ")"),
    ),
    intexp: $ => choice($.int_operation, $.num),
    num: _$ => "num",
    return_statement: $ => seq("return", $.intexp)
  }
});
