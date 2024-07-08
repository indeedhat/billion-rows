# Billion row challenge
i rewrote the whole script from scratch and removed the split lines step

this now runs fairly fast but consumes a huge amount of memory for some reason... like 20+GB

## current implementation measurement
real    0m14.626s
user    3m24.901s
sys     0m5.194s

## previous implementation measurement
DNF

## Baseline measurements
real    3m49.808s
user    3m38.623s
sys     0m10.442s

## output format
{%s=%0.1f,%0.1f,%0.1f, ...}
