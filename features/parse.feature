Feature: parse string input
  In order to generate a k-map based on user-inputted arguments
  As a program
  I need to parse said arguments

  Scenario: valid input
    Given the delimiter is ","
    When I parse the string "0,3,6,11"
    Then the parsing result should be
      | 0 | 3 | 6 | 11 |

  Scenario Outline: variable delimiters
    Given the delimiter is <delim>
    When I parse the string "<input>"
    Then the parsing result should be
      | <a> | <b> | <c> | <d> |

    Scenarios:
      | delim | input      | a | b | c | d |
      | ", "  | 8, 3, 1, 7 | 8 | 3 | 1 | 7 |
      | ":"   | 9:4:2:8    | 9 | 4 | 2 | 8 |
      | "/"   | 0/1/4/7    | 0 | 1 | 4 | 7 |
      | " "   | 3 6 2 9    | 3 | 6 | 2 | 9 |

  Scenario Outline: invalid input
    Given the delimiter is ","
    When I parse the string "<input>"
    Then an error should have occurred

    Scenarios:
      | input      |
      # Invalid delimiter
      | 0, 3, 4, 5 |
      # Non numeric character
      | 4,f,7,10   |
      # Completely invalid
      | foobar     |

  Scenario: empty input
    Given the delimiter is ""
    When I parse the string ""
    Then the parsing result should be empty

  Scenario: one argument
    Given the delimiter is ""
    When I parse the string "1"
    Then the parsing result should be
      | 1 |