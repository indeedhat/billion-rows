# Billion row challenge
parse temps as a int64 (with bastardised parseuint from strconv) and convert back to floats at the output step

## current implementation measurement
real    0m11.882s
user    1m17.480s
sys     0m6.506s

## previous implementation measurement
real    0m12.663s
user    2m25.145s
sys     0m5.452s

## Baseline measurements
real    3m49.808s
user    3m38.623s
sys     0m10.442s

## output format
{%s=%0.1f,%0.1f,%0.1f, ...}
