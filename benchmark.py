import argparse
import os.path
import subprocess
import sys


parser = argparse.ArgumentParser(description="Connect 4 benchmarker")
parser.add_argument('command')
parser.add_argument('tests', nargs="+")


def main(args):
    results = []
    with subprocess.Popen(args.command, bufsize=1, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=sys.stderr, encoding='utf-8') as solver:
        for test in args.tests:
            sys.stdout.write(test + " ")
            fh = open(test, 'r')
            totalTime = 0
            totalStates = 0
            totalLines = 0
            with open(test, 'r') as fh:
                for i, line in enumerate(fh):
                    moves, score = line.strip().split(" ", 1)
                    solver.stdin.write(moves + "\n")
                    outmoves, outscore, states, elapsed = solver.stdout.readline().split(" ", 4)
                    if moves != outmoves:
                        print("%d: Input move sequence was %s, but solver returned %s" % (i + 1, moves, outmoves))
                        sys.exit(1)
                    if score != outscore:
                        print("%d: Expected score was %s, but solver returned %s" % (i + 1, score, outscore))
                        sys.exit(1)
                    totalTime += int(elapsed)
                    totalStates += int(states)
                    totalLines += 1
                    sys.stdout.write("✓")
                    sys.stdout.flush()
            sys.stdout.write("\n")
            results.append((test, totalTime, totalStates, totalLines))
    print("Suite\t\tTime/test\tPos/test\tKStates/sec")
    for test, totalTime, totalStates, totalLines in results:
        print("%s\t%d\t%.1f\t%d" % (os.path.basename(test), totalTime / totalLines, totalStates / float(totalLines), totalStates / (float(totalTime) / 1000000) / 1000))


if __name__ == '__main__':
    args = parser.parse_args()
    main(args)
