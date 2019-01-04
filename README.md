# go4
Connect 4 solver in Go

Based on the guide [solving connect four](http://blog.gamesolver.org/solving-connect-four/).

`go4` accepts positions on standard input, formatted as a series of moves (eg, "3434") and line-delimited. It outputs space-separated 4-tuples, of the form `<moves> <score> <nstates> <elapsed>`, where `moves` is the input move sequence, `score` is the score the solver assigned, `nstates` is the number of states explored, and `elapsed` is the number of microseconds spent evaluating the position. `score` is positive if the current player will win, and negative if they will lose, and represents the number of empty spaces left. For example, `-5` means that the current player will lose with 5 empty spaces.

`benchmark` runs a series of tests, found in the `testsuite` directory, to evaluate solver performance. To run, execute `python ./path/to/go4 ./Test_Lx_Ry [...]`. Each test file contains 1000 test cases, along with expected score. Files are named after their length (more moves made already equates to an easier problem to solve) and their complexity. For instance, `Test_L3_R1` is the simplest set of tests, being long move sequences with simple solutions, while `Test_L1_R3` is the hardest, containing short sequences with difficult resolutions.
