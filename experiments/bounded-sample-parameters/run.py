#!/usr/bin/env python3
"""Evaluate Conjecture 6.1's sample bound for m=2; no encoding or timing.

Use exact integers to select the smallest t>2 with S >= 2**50.
Floating-point logarithms are used only to display the resulting bounds.
Run from any directory; results.csv is written beside this script.
"""
import csv
import math
from pathlib import Path


def sample_bound(s, t):
    """For m=2, S=min(q**(t-1), t**(t*t-1)), where q=2**s."""
    return min(2 ** (s * (t - 1)), t ** (t * t - 1))


def main():
    rows = []
    for s in range(8, 21):
        t = next(t for t in range(3, 2**s - 1)
                 if sample_bound(s, t) >= 2**50)
        # Section 7 fixes ell=q-1 and L=ell+128. Check the two
        # computational bounds of Conjecture 6.1 against 2**100.
        q = 2**s
        ell, length = q-1, q-1+128
        assert math.comb(length, ell) >= 2**100
        assert 2 * length * s >= 100  # log2(q**(m*L)), m=2
        rows.append((s, 2, t, s*(t-1),
                     f'{(t*t-1)*math.log2(t):.6f}',
                     f'{math.log2(sample_bound(s,t)):.6f}',
                     str(t in (5, 6)).lower()))
    with Path(__file__).with_name('results.csv').open('w', newline='') as f:
        writer = csv.writer(f, lineterminator="\n")
        writer.writerow(('s','m','minimum_t','birthday_log2_samples',
                         'intersection_log2_samples','log2_S','t_is_5_or_6'))
        writer.writerows(rows)


if __name__ == '__main__':
    main()
