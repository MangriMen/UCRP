#!/usr/bin/env python3
"""Smoke test: Kev-4B as a classifier over math, via /v1/systemone.

Runs only with the Python standard library. Exits non-zero if the server is
down or if any case is classified outside its expected label set.

    python3 test_math_classifier.py [--url http://127.0.0.1:9931]
"""

from __future__ import annotations

import argparse
import json
import sys
import urllib.error
import urllib.request

# Branch of math -> the label we expect the model to pick. "expected" is the
# label we want to see; "acceptable" is what would not make the test fail.
CASES = [
    {
        "name": "linear equation -> algebra",
        "state": "Solve for x: 3x + 7 = 22.",
        "expected": "algebra",
        "acceptable": {"algebra", "arithmetic"},
    },
    {
        "name": "derivative -> calculus",
        "state": "Find the derivative of f(x) = x^3 * sin(x) with respect to x.",
        "expected": "calculus",
        "acceptable": {"calculus"},
    },
    {
        "name": "triangle area -> geometry",
        "state": "A triangle has base 12 cm and height 5 cm. What is its area?",
        "expected": "geometry",
        "acceptable": {"geometry", "arithmetic"},
    },
    {
        "name": "mean of a sample -> statistics",
        "state": "The sample is 4, 8, 15, 16, 23, 42. Compute its arithmetic mean.",
        "expected": "statistics",
        "acceptable": {"statistics", "arithmetic"},
    },
    {
        "name": "prime check -> number theory",
        "state": "Prove that there are infinitely many prime numbers.",
        "expected": "number_theory",
        "acceptable": {"number_theory", "algebra"},
    },
    {
        "name": "shopping list -> not math",
        "state": "Buy milk, eggs and bread from the store on the way home.",
        "expected": "not_math",
        "acceptable": {"not_math"},
    },
    {
        "name": "is this math at all? (noul) -> yes",
        "state": "Compute the integral of 1/x from 1 to e.",
        "noul": True,
        "expected": "yes",
        "acceptable": {"yes"},
    },
    {
        "name": "is this math at all? (noul) -> no",
        "state": "The cat sat on the mat and watched the birds outside the window.",
        "noul": True,
        "expected": "no",
        "acceptable": {"no"},
    },
]

BRANCHES = {
    "algebra": "equations, expressions, polynomials, groups, rings",
    "geometry": "shapes, angles, area, volume, Euclidean constructions",
    "calculus": "limits, derivatives, integrals, series, differential equations",
    "statistics": "probability, distributions, sampling, estimation, regression",
    "number_theory": "primes, divisibility, congruences, Diophantine equations",
    "not_math": "everyday language with no mathematical content",
}


def call_systemone(url: str, state: str, *, noul: bool) -> dict:
    if noul:
        questions = {"is_math": {"type": "noul", "instructions": "Is this text about mathematics?"}}
    else:
        questions = {
            "branch": {
                "type": "choice",
                "instructions": "Which branch of mathematics does this text belong to?",
                "criteria": {k: v for k, v in BRANCHES.items()},
            }
        }
    body = json.dumps({"state": state, "questions": questions}).encode("utf-8")
    req = urllib.request.Request(
        url.rstrip("/") + "/v1/systemone",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        return json.load(resp)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--url", default="http://127.0.0.1:9931")
    args = ap.parse_args()

    print(f"target: {args.url}")
    passed = failed = 0

    for case in CASES:
        noul = bool(case.get("noul"))
        try:
            raw = call_systemone(args.url, case["state"], noul=noul)
        except (urllib.error.URLError, OSError) as exc:
            print(f"  FAIL  {case['name']}: server unreachable ({exc})")
            failed += 1
            continue

        answers = raw.get("answers", {})
        if noul:
            p_yes = float(answers.get("is_math", {}).get("noul", 0.5))
            label = "yes" if p_yes >= 0.5 else "no"
            detail = f"p(yes)={p_yes:.3f}"
        else:
            branch = answers.get("branch", {})
            label = branch.get("choice")
            probs = branch.get("probabilities", {})
            conf = branch.get("confidence")
            detail = f"choice={label} conf={conf} probs={ {k: round(v, 3) for k, v in probs.items()} }"

        ok = label in case["acceptable"]
        if ok:
            passed += 1
            mark = "PASS"
        else:
            failed += 1
            mark = "FAIL"
        print(f"  {mark}  {case['name']}: {detail}")

    print()
    print(f"result: {passed} passed, {failed} failed, {len(CASES)} total")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
