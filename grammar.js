/**
 * @file Parser grammar for tree-sitter
 * @author Agus
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

const commaSeparatedOptional = rule => optional(
  seq(repeat(seq(rule, ",")), rule)
)


export default grammar({
  name: "preprojectlang",

  rules: {
    // ────────────────────────────────────────────────────────────────────────────
    // Entry points
    // ────────────────────────────────────────────────────────────────────────────
    //source_file: ($) =>
    //  choice(seq(choice($._void_type, $._int_type, $._bool_type), $.main)),

    source_file: $ => $.program,

    program: $ => seq(
      "program",
      "{",
      repeat($.declaration_statement),
      repeat($.method_declaration_statement),
      "}"
    ),


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
      seq("{", repeat(seq(field("statement", $._statement))), "}"),

    method_call: $ =>
      seq($.identifier, "(", commaSeparatedOptional($._expression), ")"),

    _statement: ($) =>
      seq(choice(
        $.assignment_statement,
        $.method_call,
        $.return_statement,
      ), ";"),

    declaration_statement: ($) =>
      seq(field("type", $._type), field("identifier", $.identifier), "=", $._expression, ";"),

    parameter: $ => seq(
      field("type", $._type),
      field("identifier", $.identifier),
    ),

    method_declaration_statement: $ =>
      seq(
        field("type", $._type),
        field("identifier", $.identifier),
        seq("(", commaSeparatedOptional($.parameter), ")"),
        choice($.block, seq("extern", ";"))
      ),

    assignment_statement: ($) =>
      seq(
        field("identifier", $.identifier),
        "=",
        field("value", $._expression),
      ),

    return_statement: ($) =>
      seq("return", optional(field("value", $._expression))),

    // ────────────────────────────────────────────────────────────────────────────
    // Expressions
    // ────────────────────────────────────────────────────────────────────────────
    _expression: ($) => choice($._exp, seq("(", $._expression, ")")),

    _exp: ($) => prec.left(choice($._int_operation, $.num, $._bool_const, $.identifier)),

    _int_operation: ($) => choice($.int_proc, $.int_div, $.int_sum, $.int_sub),

    int_proc: ($) =>
      prec.left(
        1,
        seq(field("left", $._expression), "*", field("right", $._expression))
      ),
    int_div: ($) =>
      prec.left(
        1,
        seq(field("left", $._expression), "/", field("right", $._expression))
      ),
    int_sum: ($) =>
      prec.left(
        seq(field("left", $._expression), "+", field("right", $._expression))
      ),
    int_sub: ($) =>
      prec.left(
        seq(field("left", $._expression), "-", field("right", $._expression))
      ),

    // ────────────────────────────────────────────────────────────────────────────
    // Terminals
    // ────────────────────────────────────────────────────────────────────────────
    identifier: (_$) => /[a-z,A-Z][a-z,A-Z,0-9]*/,

    true: (_$) => "true",
    false: (_$) => "false",
    _bool_const: ($) => choice($.true, $.false),

    num: (_$) => /\d+/,
  },
});
