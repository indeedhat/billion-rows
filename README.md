# Billion row challenge
split into 3 distinct steps, read, split lines, parse

this does not work, when i let the map size of the parse step go over 423ish it runs away withe memory allocations and 
crashes... except for very occasional runs that don't for some reason...

i think im making it too complicated, maybe i step back to fewer stages and combine the split lines/parse steps

## current implementation measurement
DNF

## previous implementation measurement
real    1m43.852s
user    5m0.222s
sys     0m41.819s

## Baseline measurements
real    3m49.808s
user    3m38.623s
sys     0m10.442s

## output format
{%s=%0.1f,%0.1f,%0.1f, ...}
