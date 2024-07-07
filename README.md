# Billion row challenge
using a more sensible approach of batching input lines to send to the go routines

## current implementation measurement
real    2m9.149s
user    6m46.196s
sys     0m49.941s

## Baseline measurements
real    3m49.808s
user    3m38.623s
sys     0m10.442s

## output format
{%s=%0.1f,%0.1f,%0.1f, ...}
