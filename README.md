# Billion row challenge
split reading and parsing into seperate goroutines

I suspected this might be a little slower due to the inclusion of a poorly optimized chanel but i 
wasn't expecting such an impact

## current implementation measurement
real    6m46.305s
user    18m35.770s
sys     2m54.213s

## Baseline measurements
real    3m49.808s
user    3m38.623s
sys     0m10.442s

## output format
{%s=%0.1f,%0.1f,%0.1f, ...}
