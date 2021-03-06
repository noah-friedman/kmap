Feature: command line arguments
  In order to have users easily provide program arguments
  As a program
  I need to parse command line arguments

  Scenario Outline: -size, -s
    Given I set the "<arg>" command-line argument to "<size>"
    When I run the program
    Then I should be asked "What are the arguments to the k-map?:"
    When I answer ""
    Then I should be asked "What are the don't care conditions of the k-map?:"
    When I answer ""
    Then the program should output an empty k-map of size <size>
    And the program should exit cleanly

    Scenarios:
      | arg  | size |
      | size | 2    |
      | size | 3    |
      | size | 4    |
      | s    | 2    |
      | s    | 3    |
      | s    | 4    |

  Scenario Outline: -args, -a
    Given I set the "<arg>" command-line argument to "1,3,7"
    When I run the program
    Then I should be asked "What is the size of the k-map? (3):"
    When I answer ""
    Then I should be asked "What are the don't care conditions of the k-map?:"
    When I answer ""
    Then the program should output
      """
                    y
        -----------------
        | 0 | 1 | 1 | 0 |
        -----------------
      x | 0 | 0 | 1 | 0 |
        -----------------
                z
      """
    And the program should exit cleanly

    Scenarios:
      | arg  |
      | args |
      | a    |

  Scenario Outline: -dont-care, -dc
    Given I set the "<arg>" command-line argument to "4,5,7"
    When I run the program
    Then I should be asked "What is the size of the k-map? (3):"
    When I answer ""
    Then I should be asked "What are the arguments to the k-map?:"
    When I answer ""
    Then the program should output
      """
                    y
        -----------------
        | 0 | 0 | 0 | 0 |
        -----------------
      x | X | X | X | 0 |
        -----------------
                z
      """

    Scenarios:
      | arg       |
      | dont-care |
      | dc        |