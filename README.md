# Billion row challenge
- perform string conversion on chunk before passing to chunk channel to optimize memcopy
- fix bug where i was creting a buffer 1 byte too large to deal with chunk overflow

## current implementation measurement
real    0m6.805s
user    1m39.661s
sys     0m2.179s

## previous implementation measurement
real    0m7.620s
user    1m34.781s
sys     0m4.370s

## Baseline measurements
real    3m49.808s
user    3m38.623s
sys     0m10.442s

## output format
{%s=%0.1f,%0.1f,%0.1f, ...}
