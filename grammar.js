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
    // TODO: add the actual grammar rules
    source_file: $ => choice(seq("void", $.main)),
    main: $ => seq("main", $._args, $._block),
    _args: _$ => seq("(", repeat("arg"), ")"),
    _block: $ => seq("{", repeat(seq($._statement, ";")), "}"),
    _statement: $ => (choice($.return_statement, "skip", $.intexp)),
    intexp: $ => choice((seq("(", $._int_operation, ")"), $._num)),
    _int_operation: $ => choice(
      seq($._num, "+", $._num),
      seq($._num, "/", $._num),
      seq($._num, "*", $._num),
      seq($._num, "-", $._num),
    ),
    _num: _$ => "num",
    return_statement: $ => seq("return", $.intexp)
  }
});
